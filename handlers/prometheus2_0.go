package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/circonus-labs/gosnowth"
	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/prompb"
	uuid "github.com/satori/go.uuid"
)

type promRequestParams struct {
	accountID int32
	checkName string
	checkUUID uuid.UUID
}

func extractPromRequest(ctx echo.Context, req proto.Message) (promRequestParams, error) {
	fullstart := time.Now()
	var (
		// create our prometheus format decoder
		err error
		prp = promRequestParams{}
	)

	// validation of url parameters
	prp.accountID, err = ValidateAccountID(ctx)
	if err != nil {
		// 400, invalid account id
		return prp, echo.NewHTTPError(http.StatusBadRequest, "invalid account id in URL")
	}
	prp.checkUUID, err = ValidateCheckUUID(ctx)
	if err != nil {
		// 400, invalid account id
		return prp, echo.NewHTTPError(http.StatusBadRequest, "invalid check_uuid in URL")
	}
	prp.checkName, err = ValidateCheckName(ctx)
	if err != nil {
		// 400, invalid account id
		return prp, echo.NewHTTPError(http.StatusBadRequest, "invalid check_name in URL")
	}

	if ctx.Request().Body == nil {
		// 400, invalid account id
		return prp, echo.NewHTTPError(http.StatusBadRequest, "request requires a body")
	}
	defer ctx.Request().Body.Close()

	start := time.Now()
	// pull the body off of the request into a byte slice
	compressed, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.Logger().Errorf("failed to read request body: %s", err.Error())
		return prp, echo.NewHTTPError(http.StatusBadRequest, "could not read request body")
	}
	ctx.Logger().Warnf("timing - read request body: %+v", time.Now().Sub(start))

	start = time.Now()
	reqBuf, err := snappy.Decode(nil, compressed)
	if err != nil {
		ctx.Logger().Errorf("failed to decompress request body: %s", err.Error())
		return prp, echo.NewHTTPError(http.StatusBadRequest, "failed to decode request body")
	}
	ctx.Logger().Warnf("timing - snappy decode: %+v", time.Now().Sub(start))

	start = time.Now()
	if err := proto.Unmarshal(reqBuf, req); err != nil {
		ctx.Logger().Errorf("failed to decode protobuf request: %s", err.Error())
		return prp, echo.NewHTTPError(http.StatusBadRequest, "failed to decode request body")
	}
	ctx.Logger().Warnf("timing - read protobuf: %+v", time.Now().Sub(start))
	ctx.Logger().Warnf("timing - extract prom request: %+v", time.Now().Sub(fullstart))
	return prp, nil
}

// PrometheusWrite2_0 - the application handler which converts a prometheus
// MetricFamily message to a MetricList for ingestion into IRONdb
func PrometheusWrite2_0(ctx echo.Context) error {
	// decode, read and deserialize the prometheus request
	var (
		req          = new(prompb.WriteRequest)
		prp          promRequestParams
		err          error
		snowthClient SnowthClientI
	)

	if client, ok := ctx.Get("snowthClient").(SnowthClientI); ok {
		snowthClient = client
	} else {
		ctx.Logger().Errorf("no snowth client in context %+v\n", ctx)
		return errors.New("no active snowth nodes")
	}

	if prp, err = extractPromRequest(ctx, req); err != nil {
		return err
	}

	// make the metric list flatbuffer data
	metricList, err := MakeMetricList(
		req.GetTimeseries(), prp.accountID, prp.checkName, prp.checkUUID)
	if err != nil {
		ctx.Logger().Errorf("failed to convert to flatbuffer: %s", err.Error())
		return err
	}

	// pull a random snowth node from the client to send request to
	node := ChooseActiveNode(snowthClient)
	if node == nil {
		// we are degraded, there are no active nodes
		ctx.Logger().Errorf("there are no active nodes... active: %+v, inactive: %+v\n", snowthClient.ListActiveNodes(), snowthClient.ListInactiveNodes())
		return errors.New("no active irondb nodes")
	}

	ctx.Logger().Debugf(
		"using node: %s of %+v",
		node.GetURL(), snowthClient.ListActiveNodes())

	// perform the write to IRONdb
	if err = snowthClient.WriteRaw(node, bytes.NewBuffer(metricList), true, uint64(len(req.GetTimeseries()))); err != nil {
		id := uuid.NewV4()
		if strings.Contains(err.Error(), "Bad Request") || strings.Contains(err.Error(), "400") {
			if err := ioutil.WriteFile("/tmp/irondb-prometheus-adapter_"+id.String(), metricList, 0644); err != nil {
				ctx.Logger().Warnf("failed to write metric list, data file: %+v", err)
			}
			ctx.Logger().Errorf(
				"failed to write flatbuffer: metriclist written -> /tmp/irondb-prometheus-adapter_%s -> %+v",
				id, id, err)
		} else {
			ctx.Logger().Errorf("failed to write to snowth /raw -> %+v", err)
		}
		return errors.Wrap(err, "failed to write flatbuffer")
	}
	return nil
}

// PrometheusRead2_0 - the application handler which converts a prometheus
// read message to an IRONdb read message, and returns the results converted
// back into prometheus output
func PrometheusRead2_0(ctx echo.Context) error {
	// decode, read and deserialize the prometheus request
	var (
		req          = new(prompb.ReadRequest)
		resp         = new(prompb.ReadResponse)
		prp          promRequestParams
		err          error
		snowthClient SnowthClientI
	)
	if client, ok := ctx.Get("snowthClient").(SnowthClientI); ok {
		snowthClient = client
	}

	start := time.Now()
	if prp, err = extractPromRequest(ctx, req); err != nil {
		return err
	}

	ctx.Logger().Debugf("handlePromRequest duration: %v\n", time.Now().Sub(start))

	// pull a random snowth node from the client to send request to
	node := ChooseActiveNode(snowthClient)
	if node == nil {
		// we are degraded, there are no active nodes
		ctx.Logger().Errorf("there are no active nodes... active: %+v, inactive: %+v\n", snowthClient.ListActiveNodes(), snowthClient.ListInactiveNodes())
		return errors.New("no active irondb nodes")
	}
	ctx.Logger().Debugf("using node: %s of %+v", node.GetURL(), snowthClient.ListActiveNodes())
	// convert the read message into a tags snowth query
	// perform the tags snowth query
	// take all resulting metrics, and perform a time bound metrics query
	// convert the results into the response
	ctx.Logger().Warnf("prometheus query: %+v", req.GetQueries())

	start = time.Now()
	for _, q := range req.GetQueries() {
		// foreach query, perform the query and generate a query result
		var (
			// for each query we will be making a queryresponse
			qr             = new(prompb.QueryResult)
			snowthTagQuery strings.Builder
			streamTags     = []string{}
		)

		// always include the check_uuid in the tag query, will reduce search space
		snowthTagQuery.WriteString("and(__check_uuid:")
		snowthTagQuery.WriteString(prp.checkUUID.String())
		snowthTagQuery.WriteString(",")
		for i, m := range q.GetMatchers() {
			// for each of the matchers within the query
			// take each matcher and formulate a stream tag filter
			if i > 0 {
				snowthTagQuery.WriteByte(',')
			}

			var (
				name  string = m.GetName()
				value string = m.GetValue()
			)
			if name == "__name__" {
				name = "__name"
			}

			var (
				matcherName = base64.StdEncoding.EncodeToString(
					[]byte(name))
				matcherValue = base64.StdEncoding.EncodeToString(
					[]byte(value))
			)

			// based on the matcher type, we need to build out our tag query
			switch m.Type {
			case prompb.LabelMatcher_EQ:
				// query equal
				tag := fmt.Sprintf(`b"%s":b"%s"`, matcherName, matcherValue)
				snowthTagQuery.WriteString(tag)
				streamTags = append(streamTags, tag)
			case prompb.LabelMatcher_NEQ:
				// query not equal
				tag := fmt.Sprintf(`b"%s":b"%s"`, matcherName, matcherValue)
				snowthTagQuery.WriteString("not(")
				snowthTagQuery.WriteString(tag)
				snowthTagQuery.WriteByte(')')
			case prompb.LabelMatcher_RE:
				// query regular expression
				tag := fmt.Sprintf(`b"%s":b/%s/`, matcherName, matcherValue)
				snowthTagQuery.WriteString(tag)
				streamTags = append(streamTags, tag)
			case prompb.LabelMatcher_NRE:
				// query not regular expression
				tag := fmt.Sprintf(`b"%s":b/%s/`, matcherName, matcherValue)
				snowthTagQuery.WriteString("not(")
				snowthTagQuery.WriteString(tag)
				snowthTagQuery.WriteByte(')')
			}
		}
		// close our and(
		snowthTagQuery.WriteByte(')')

		var (
			tagResp []gosnowth.FindTagsItem
			err     error
		)

		start := time.Now()
		ctx.Logger().Warnf("timing find query: %s", snowthTagQuery.String())
		tagResp, err = snowthClient.FindTags(node, prp.accountID, snowthTagQuery.String(), "", "")
		ctx.Logger().Warnf("timing find query: %s, duration: %v", snowthTagQuery.String(), time.Now().Sub(start))

		if err != nil {
			ctx.Logger().Errorf("failed to find tags: %s", err.Error())
			continue
		}

		ctx.Logger().Warnf("doing %d rollups for query: %s", len(tagResp), snowthTagQuery.String())

		var tsChan = make(chan *prompb.TimeSeries, 100)

		var step = 60 * time.Second
		if q.Hints.StepMs > 0 {
			step = time.Duration(q.Hints.StepMs) * time.Millisecond
		}

		for _, v := range tagResp {
			go func() {
				// for all of our tag responses, grab the rollups
				start := time.Now()
				values, err := snowthClient.ReadRollupValues(
					node, prp.checkUUID.String(), v.MetricName, []string{}, step,
					time.Unix(0, q.StartTimestampMs*int64(time.Millisecond)),
					time.Unix(0, q.EndTimestampMs*int64(time.Millisecond)),
				)

				ctx.Logger().Warnf("timing rollup query: %s, result length: %d, duration: %v", v.MetricName, len(values), time.Now().Sub(start))
				ctx.Logger().Debugf("rollup results: %+v", values)
				timeSeries := new(prompb.TimeSeries)
				if err != nil {
					ctx.Logger().Errorf("failed to read rollup: %s", err.Error())
					tsChan <- timeSeries
					return
				}

				timeSeries.Labels = metricNameToLabelPairs(v.MetricName)

				for _, v := range values {
					// convert value to time series
					timeSeries.Samples = append(timeSeries.Samples,
						&prompb.Sample{
							Value:     v.Value,
							Timestamp: v.Timestamp * 1000, // prom does ms
						})
				}
				// add the timeseries to the query result
				tsChan <- timeSeries
			}()
		}

		for timeSeries := range tsChan {
			ctx.Logger().Warnf("time series added to resultset, #samples: %d", len(timeSeries.Samples))
			qr.Timeseries = append(qr.Timeseries, timeSeries)
		}
		// add this result to our results
		resp.Results = append(resp.Results, qr)
	}
	ctx.Logger().Warnf("total results, #query-resp: %d", len(resp.Results))

	ctx.Logger().Warnf("total read duration: %v", time.Now().Sub(start))
	ctx.Logger().Debugf("results: ", resp)

	start = time.Now()
	data, err := proto.Marshal(resp)
	if err != nil {
		ctx.Logger().Errorf("failed to marshal response: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to marshal response")
	}

	ctx.Response().Header().Set("Content-Type", "application/x-protobuf")
	ctx.Response().Header().Set("Content-Encoding", "snappy")
	ctx.Response().WriteHeader(http.StatusOK)

	var compressed = snappy.Encode(nil, data)
	if _, err := ctx.Response().Write(compressed); err != nil {
		ctx.Logger().Errorf("failed to write response: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to write response")
	}
	ctx.Logger().Warnf("total read handler timing: %v", time.Now().Sub(start))
	return nil
}

var canonicalMetricNameRE = regexp.MustCompile(`^(.+)\|ST\[(.+)\]$`)

func metricNameToLabelPairs(metricName string) []*prompb.Label {
	// up|ST[cmdb_shard:test,cmdb_status:ALLOCATED,data_center:dal06,environment:prod,instance:prometheus-test-000-g5.prod.dal06.fitbit.com:9201,job:prometheus-remote-storage,monitor:test,replica:prometheus-test-000-g5.prod.dal06.fitbit.com,tier:prometheus]
	if !canonicalMetricNameRE.MatchString(metricName) {
		return []*prompb.Label{
			&prompb.Label{Name: "__name__", Value: metricName},
		}
	}
	metricNameParts := canonicalMetricNameRE.FindAllStringSubmatch(metricName, -1)
	if len(metricNameParts) < 1 || len(metricNameParts[0]) < 3 {
		return []*prompb.Label{
			&prompb.Label{
				Name: "__name__", Value: metricName,
			},
		}
	}

	labelPairs := []*prompb.Label{
		&prompb.Label{
			Name: "__name__", Value: metricNameParts[0][1],
		},
	}

	for _, v := range strings.Split(metricNameParts[0][2], ",") {
		vv := strings.Split(v, ":")
		if len(vv) > 1 {
			labelPairs = append(labelPairs, &prompb.Label{
				Name: vv[0], Value: strings.Join(vv[1:], ":"),
			})
		}
	}
	return labelPairs
}
