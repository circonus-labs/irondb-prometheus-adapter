package handlers

import (
	"encoding/base64"
	"fmt"
	"strconv"

	circfb "github.com/circonus-labs/irondb-prometheus-adapter/flatbuffer/circonus"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/pkg/errors"
	dto "github.com/prometheus/client_model/go"
)

// MakeMetricList - given a prometheus MetricFamily pointer, an
// accountID, checkName and check UUID, generate the Flatbuffer
// serialized message that is to be sent to the IRONdb api.
// This will result in a []byte and error, where the []byte is
// the corresponding MetricList serialized
func MakeMetricList(promMetricFamily *dto.MetricFamily,
	accountID, checkName, checkUUID string) ([]byte, error) {
	var (
		// b is the flatbuffer builder used to create the MetricList
		b = flatbuffers.NewBuilder(0)
		// offsets is a list of metric offsets used to build the MetricList
		offsets = []flatbuffers.UOffsetT{}
	)

	// convert metric and labels to IRONdb format:
	for _, metric := range promMetricFamily.Metric {
		// MakeMetric takes in a flatbuffer builder, the metric from
		// the prometheus metric family and results in an offset for
		// the metric inserted into the builder
		mOffset, err := MakeMetric(b, metric, accountID, checkName, checkUUID)
		if err != nil {
			return []byte{}, errors.Wrap(err,
				"failed to encode metric to flatbuffer")
		}
		// keep track of all of the metric offsets so we can build the
		// MetricList Metrics Vector
		offsets = append(offsets, mOffset)
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
	b.Finish(metricListOffset)

	// return the finished serialized bytes
	return b.FinishedBytes(), nil
}

// MakeMetric - serialize a prometheus Metric as a flatbuffer resulting
// in the offset on the builder for the Metric
func MakeMetric(b *flatbuffers.Builder, promMetric *dto.Metric,
	accountID, checkName, checkUUID string) (flatbuffers.UOffsetT, error) {
	var (
		// apply the checkName and UUID to the metric
		checkNameOffset = b.CreateString(checkName)
		checkUUIDOffset = b.CreateString(checkUUID)
		tagOffsets      = []flatbuffers.UOffsetT{}
	)
	// we need to convert the labels into stream tag format
	for _, labelPair := range promMetric.GetLabel() {
		tagOffsets = append(tagOffsets, b.CreateString(fmt.Sprintf(`b"%s":b"%s"`,
			labelPair.GetName(),
			base64.StdEncoding.EncodeToString([]byte(labelPair.GetValue())))))
	}
	circfb.MetricValueStartStreamTagsVector(b, len(promMetric.GetLabel()))
	for _, offset := range tagOffsets {
		b.PrependUOffsetT(offset)
	}
	streamTagVec := b.EndVector(len(promMetric.GetLabel()))

	// create the metric value
	circfb.MetricValueStart(b)
	// add timestamp to metric value
	circfb.MetricValueAddTimestamp(b, uint64(promMetric.GetTimestampMs()))
	// add name to metric value
	circfb.MetricValueAddName(b, checkNameOffset)
	circfb.MetricValueAddStreamTags(b, streamTagVec)
	value := circfb.MetricValueEnd(b)

	// start a metric
	circfb.MetricStart(b)
	// add the timestamp
	circfb.MetricAddTimestamp(b, uint64(promMetric.GetTimestampMs()))
	// add the check name
	circfb.MetricAddCheckName(b, checkNameOffset)
	// add the check uuid
	circfb.MetricAddCheckUuid(b, checkUUIDOffset)
	// add the account id
	aid, err := strconv.ParseInt(accountID, 10, 32)
	if err != nil {
		return 0, errors.Wrap(err, "failed to convert account id")
	}
	// add the account ID to the Metric
	circfb.MetricAddAccountId(b, int32(aid))
	circfb.MetricAddValue(b, value)
	metric := circfb.MetricEnd(b)
	// return the offset of the metric to the caller
	return metric, nil
}
