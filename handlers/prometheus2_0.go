package handlers

import (
	"bytes"
	"net/http"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// PrometheusWrite2_0 - the application handler which converts a prometheus
// MetricFamily message to a MetricList for ingestion into IRONdb
func PrometheusWrite2_0(ctx echo.Context) error {
	var (
		// create our prometheus format decoder
		dec          = expfmt.NewDecoder(ctx.Request().Body, expfmt.Format(ctx.Request().Header.Get("Content-Type")))
		metricFamily = new(dto.MetricFamily)
		err          error
		data         []byte
		snowthClient = ctx.Get("snowthClient").(SnowthClientI)
	)

	// validation of url parameters
	accountID, err := ValidateAccountID(ctx)
	if err != nil {
		// 400, invalid account id
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account id in URL")
	}
	checkUUID, err := ValidateCheckUUID(ctx)
	if err != nil {
		// 400, invalid account id
		return echo.NewHTTPError(http.StatusBadRequest, "invalid check_uuid in URL")
	}
	checkName, err := ValidateCheckName(ctx)
	if err != nil {
		// 400, invalid account id
		return echo.NewHTTPError(http.StatusBadRequest, "invalid check_name in URL")
	}

	// close request body
	defer ctx.Request().Body.Close()

	// decode the metrics into the metric family
	err = dec.Decode(metricFamily)
	if err != nil {
		ctx.Logger().Errorf("failed to decode: %s", err.Error())
		return err
	}
	ctx.Logger().Debugf("parsed metric-family: %+v\n", metricFamily)

	// make the metric list flatbuffer data
	data, err = MakeMetricList(metricFamily, accountID, checkName, checkUUID)
	if err != nil {
		ctx.Logger().Errorf("failed to convert to flatbuffer: %s", err.Error())
		return err
	}

	// pull a random snowth node from the client to send request to
	node := getRandomNode(snowthClient.ListActiveNodes()...)
	if node == nil {
		// we are degraded, there are no active nodes
		ctx.Logger().Errorf("there are no active nodes... active: %+v, inactive: %+v\n", snowthClient.ListActiveNodes(), snowthClient.ListInactiveNodes())
		return errors.New("no active irondb nodes")
	}
	ctx.Logger().Debugf("using node: %s of %+v", node.GetURL(), snowthClient.ListActiveNodes())
	// perform the write to IRONdb
	if err = snowthClient.WriteRaw(node, bytes.NewBuffer(data), true, uint64(len(metricFamily.Metric))); err != nil {
		ctx.Logger().Errorf("failed to write flatbuffer: %s", err.Error())
		return errors.Wrap(err, "failed to write flatbuffer")
	}
	ctx.Logger().Debugf("converted flatbuffer: %+v\n", data)
	return nil
}

// PrometheusRead2_0 - the application handler which converts a prometheus
// read message to an IRONdb read message, and returns the results converted
// back into prometheus output
func PrometheusRead2_0(ctx echo.Context) error {
	return nil
}
