package handlers

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	circfb "github.com/circonus-labs/irondb-prometheus-adapter/flatbuffer/metrics"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/prompb"
	uuid "github.com/satori/go.uuid"
)

// MakeMetricList - given a prometheus MetricFamily pointer, an
// accountID, checkName and check UUID, generate the Flatbuffer
// serialized message that is to be sent to the IRONdb api.
// This will result in a []byte and error, where the []byte is
// the corresponding MetricList serialized
func MakeMetricList(timeseries []*prompb.TimeSeries,
	accountID int32, checkName string, checkUUID uuid.UUID) ([]byte, error) {
	var (
		// b is the flatbuffer builder used to create the MetricList
		b = flatbuffers.NewBuilder(0)
		// offsets is a list of metric offsets used to build the MetricList
		offsets = []flatbuffers.UOffsetT{}
	)

	for _, ts := range timeseries {
		// convert metric and labels to IRONdb format:
		for _, sample := range ts.GetSamples() {
			// MakeMetric takes in a flatbuffer builder, the metric from
			// the prometheus metric family and results in an offset for
			// the metric inserted into the builder
			mOffset, err := MakeMetric(b, ts.GetLabels(), sample, accountID, checkName, checkUUID)
			if err != nil {
				return []byte{}, errors.Wrap(err,
					"failed to encode metric to flatbuffer")
			}
			// keep track of all of the metric offsets so we can build the
			// MetricList Metrics Vector
			offsets = append(offsets, mOffset)
		}
	}
	// create a metrics vector
	circfb.MetricListStartMetricsVector(b, len(offsets))
	for _, offset := range offsets {
		// add all of the metric offsets to the vector
		b.PrependUOffsetT(offset)

	}
	// finish building the vector
	metricsVec := b.EndVector(len(offsets))

	// start the main MetricList
	circfb.MetricListStart(b)
	// add our metricsVector to the MetricList
	circfb.MetricListAddMetrics(b, metricsVec)
	var metricListOffset = circfb.MetricListEnd(b)

	b.Prep(flatbuffers.SizeInt32, 0)
	b.PlaceInt32(0)
	b.FinishWithFileIdentifier(metricListOffset, []byte("CIML"))
	// return the finished serialized bytes
	return b.FinishedBytes(), nil
}

// MakeMetric - serialize a prometheus Metric as a flatbuffer resulting
// in the offset on the builder for the Metric
func MakeMetric(b *flatbuffers.Builder, labels []*prompb.Label, sample *prompb.Sample,
	accountID int32, checkName string, checkUUID uuid.UUID) (flatbuffers.UOffsetT, error) {

	// prometheus metric types are as follows:
	// MetricType_COUNTER   MetricType = 0 -> NNT
	// MetricType_GAUGE     MetricType = 1 -> NNT
	// MetricType_SUMMARY   MetricType = 2 -> histogram
	// MetricType_UNTYPED   MetricType = 3 -> NNT/text?
	// MetricType_HISTOGRAM MetricType = 4 -> histogram

	var (
		// apply the checkName and UUID to the metric
		metricName      = ""
		checkNameOffset = b.CreateString(checkName)
		checkUUIDOffset = b.CreateString(checkUUID.String())
		tagOffsets      = []flatbuffers.UOffsetT{}
		STReprBuilder   strings.Builder
	)

	STReprBuilder.WriteString("|ST[")
	// we need to convert the labels into stream tag format
	for i, label := range labels {
		if label.GetName() == "__name__" {
			metricName = label.GetValue()
		}
		if i != 0 {
			STReprBuilder.WriteByte(',')
		}

		pair := fmt.Sprintf(`b"%s":b"%s"`,
			base64.StdEncoding.EncodeToString([]byte(label.GetName())),
			base64.StdEncoding.EncodeToString([]byte(label.GetValue())))

		STReprBuilder.WriteString(pair)

		tagOffsets = append(tagOffsets, b.CreateString(pair))
	}
	STReprBuilder.WriteByte(']')

	metricNameOffset := b.CreateString(metricName + STReprBuilder.String())
	circfb.MetricValueStartStreamTagsVector(b, len(labels))
	for _, offset := range tagOffsets {
		b.PrependUOffsetT(offset)
	}
	streamTagVec := b.EndVector(len(labels))

	// TODO: if metric type is counter/gauge do the below,
	// if histogram/summary we need to use those union types.

	// create the metric value value
	circfb.DoubleValueStart(b)
	circfb.DoubleValueAddValue(b, sample.GetValue())
	valueValue := circfb.DoubleValueEnd(b)

	// create the metric value
	circfb.MetricValueStart(b)
	// add timestamp to metric value
	var timestamp = uint64(sample.GetTimestamp())
	if timestamp == 0 {
		// not here, we should add a timestamp
		timestamp = uint64(time.Now().UnixNano() / int64(time.Millisecond))
	}

	circfb.MetricValueAddTimestamp(b, timestamp)
	// add name to metric value
	circfb.MetricValueAddName(b, metricNameOffset)
	circfb.MetricValueAddStreamTags(b, streamTagVec)
	// this is the value of the value...
	circfb.MetricValueAddValueType(b, circfb.MetricValueUnionDoubleValue)
	circfb.MetricValueAddValue(b, valueValue)

	value := circfb.MetricValueEnd(b)
	// start a metric
	circfb.MetricStart(b)
	// add the check name
	circfb.MetricAddCheckName(b, checkNameOffset)
	// add the check uuid
	circfb.MetricAddCheckUuid(b, checkUUIDOffset)
	// add the account ID to the Metric
	circfb.MetricAddAccountId(b, accountID)
	circfb.MetricAddValue(b, value)
	circfb.MetricAddTimestamp(b, timestamp)
	fid := []byte("CIMM")
	// alignment...
	b.Prep(4, 0)
	for i := 4 - 1; i >= 0; i-- {
		// place the file identifier
		b.PlaceByte(fid[i])
	}
	metric := circfb.MetricEnd(b)

	// return the offset of the metric to the caller
	return metric, nil
}
