package handlers

import (
	"strconv"

	circfb "github.com/circonus/promadapter/flatbuffer/circonus"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func MakeMetric(b *flatbuffers.Builder, promMetric *dto.Metric,
	accountID, checkName, checkUuid string) (flatbuffers.UOffsetT, error) {
	// start a new metric

	var (
		checkNameOffset = b.CreateString(checkName)
		checkUuidOffset = b.CreateString(checkUuid)
	)

	circfb.MetricStart(b)
	// add the timestamp
	circfb.MetricAddTimestamp(b, uint64(promMetric.GetTimestampMs()))
	// add the check name
	circfb.MetricAddCheckName(b, checkNameOffset)
	// add the check uuid
	circfb.MetricAddCheckUuid(b, checkUuidOffset)
	// add the account id
	aid, err := strconv.ParseInt(accountID, 10, 32)
	if err != nil {
		return 0, errors.Wrap(err, "failed to convert account id")
	}
	circfb.MetricAddAccountId(b, int32(aid))
	metric := circfb.MetricEnd(b)
	return metric, nil
}

func PrometheusWrite2_0(ctx echo.Context) error {
	var (
		// create our prometheus format decoder
		dec          = expfmt.NewDecoder(ctx.Request().Body, expfmt.Format(ctx.Request().Header.Get("Content-Type")))
		metricFamily = new(dto.MetricFamily)
		b            = flatbuffers.NewBuilder(0)
		err          error
	)
	// close request body
	defer ctx.Request().Body.Close()

	// decode the metrics into the metric family
	err = dec.Decode(metricFamily)
	if err != nil {
		ctx.Logger().Errorf("failed to decode: %s", err.Error())
		return err
	}
	ctx.Logger().Debugf("parsed metric-family: %+v\n", metricFamily)

	offsets := []flatbuffers.UOffsetT{}
	// convert metric and labels to IRONdb format:
	for _, metric := range metricFamily.Metric {
		ctx.Logger().Debugf("metric: %+v\n", metric)
		metricOffset, err := MakeMetric(b, metric, ctx.Param("account"),
			ctx.Param("check_name"), ctx.Param("check_uuid"))
		if err != nil {
			// error encoding to flatbuffer
			return errors.Wrap(err, "failed to encode metric to flatbuffer")
		}
		offsets = append(offsets, metricOffset)
	}

	circfb.MetricListStart(b)
	for _, offset := range offsets {
		circfb.MetricListAddMetrics(b, offset)
	}
	metricsList := circfb.MetricListEnd(b)
	b.Finish(metricsList)

	ctx.Logger().Debugf("output of the flatbuffer: %+v\n", b.FinishedBytes())

	return nil
}

func PrometheusRead2_0(ctx echo.Context) error {
	return nil
}
