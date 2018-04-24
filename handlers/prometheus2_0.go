package handlers

import (
	"strconv"

	circfb "github.com/circonus-labs/irondb-prometheus-adapter/flatbuffer/circonus"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func MakeMetricList(promMetricFamily *dto.MetricFamily,
	accountID, checkName, checkUUID string) ([]byte, error) {
	var (
		b = flatbuffers.NewBuilder(0)
	)

	offsets := []flatbuffers.UOffsetT{}
	// convert metric and labels to IRONdb format:
	for _, metric := range promMetricFamily.Metric {
		mOffset, err := MakeMetric(b, metric, accountID, checkName, checkUUID)
		if err != nil {
			return []byte{}, errors.Wrap(err,
				"failed to encode metric to flatbuffer")
		}
		offsets = append(offsets, mOffset)
	}
	circfb.MetricListStartMetricsVector(b, len(offsets))
	for _, offset := range offsets {
		b.PrependUOffsetT(offset)
	}
	metricsVec := b.EndVector(len(offsets))

	circfb.MetricListStart(b)
	circfb.MetricListAddMetrics(b, metricsVec)
	var metricListOffset = circfb.MetricListEnd(b)
	b.Finish(metricListOffset)

	return b.FinishedBytes(), nil
}

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
		err          error
		data         []byte
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

	data, err = MakeMetricList(metricFamily, ctx.Param("account"),
		ctx.Param("check_name"), ctx.Param("check_uuid"))
	if err != nil {
		ctx.Logger().Errorf("failed to convert to flatbuffer: %s", err.Error())
		return err
	}

	// call snowth with flatbuffer data
	ctx.Logger().Debugf("converted flatbuffer: %+v\n", data)

	return nil
}

func PrometheusRead2_0(ctx echo.Context) error {
	return nil
}
