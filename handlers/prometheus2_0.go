package handlers

import (
	"errors"
	"fmt"

	"github.com/labstack/echo"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func PrometheusWrite2_0(ctx echo.Context) error {
	var (
		// create our prometheus format decoder
		dec          = expfmt.NewDecoder(ctx.Request().Body, expfmt.Format(ctx.Request().Header.Get("Content-Type")))
		metricFamily = new(dto.MetricFamily)
		//		b            = flatbuffers.NewBuilder(0)
		err error
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
	// convert metric and labels to IRONdb format:
	for _, metric := range metricFamily.Metric {
		ctx.Logger().Debugf("metric: %+v\n", metric)

		var (
			timestamp = fmt.Sprintf("%d.%03d", metric.GetTimestampMs()/1000, metric.GetTimestampMs()%1000)
			// TODO: figure out, must need to go to DB for this??
			uuid       = "TARGET`MODULE`CIRCONUS_NAME`lower-cased-uuid"
			name       = metricFamily.GetName()
			prefix     string
			rawFmt     = "%s\t%s\t%s\t%s\t%s\t%f\n"
			metricType = "n"
			value      float64
		)

		// set metric type
		// set value
		// set format string
		switch metricFamily.GetType() {
		case dto.MetricType_COUNTER:
			value = metric.GetCounter().GetValue()
			prefix = "M"
			break
		case dto.MetricType_GAUGE:
			prefix = "M"
			value = metric.GetGauge().GetValue()
			break
			// TODO: what is a summary
			//	case dto.MetricType_SUMMARY:
			//		prefix = "M"
			//		value = metric.GetSummary().
			//		break
		case dto.MetricType_UNTYPED:
			prefix = "M"
			value = metric.GetUntyped().GetValue()
			break
		case dto.MetricType_HISTOGRAM:
			prefix = "H1"
			// TODO: convert all these buckets to a circ-hist
			//value = metric.GetHistogram().
			break
		default:
			return errors.New("invalid metric family type")
		}
		ctx.Logger().Debugf(rawFmt, prefix, timestamp, uuid, name, metricType, value)
	}

	return nil
}

func PrometheusRead2_0(ctx echo.Context) error {
	return nil
}
