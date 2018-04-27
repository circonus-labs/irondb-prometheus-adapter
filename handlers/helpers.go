package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/circonus-labs/gosnowth"
	circfb "github.com/circonus-labs/irondb-prometheus-adapter/flatbuffer/metrics"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/pkg/errors"
	dto "github.com/prometheus/client_model/go"
	uuid "github.com/satori/go.uuid"
)

type SnowthClientI interface {
	WriteRaw(*gosnowth.SnowthNode, io.Reader, bool, uint64) error
	ListActiveNodes() []*gosnowth.SnowthNode
	ListInactiveNodes() []*gosnowth.SnowthNode
}

type mockSnowthClient struct {
	mockWriteRaw          func(*gosnowth.SnowthNode, io.Reader, bool, uint64) error
	mockListActiveNodes   func() []*gosnowth.SnowthNode
	mockListInactiveNodes func() []*gosnowth.SnowthNode
}

func (msc *mockSnowthClient) WriteRaw(node *gosnowth.SnowthNode, data io.Reader, fb bool, numDatapoints uint64) (err error) {
	if msc.mockWriteRaw != nil {
		return msc.mockWriteRaw(node, data, fb, numDatapoints)
	}
	return nil
}

func (msc *mockSnowthClient) ListActiveNodes() []*gosnowth.SnowthNode {
	if msc.mockListActiveNodes != nil {
		return msc.mockListActiveNodes()
	}
	return []*gosnowth.SnowthNode{new(gosnowth.SnowthNode)}
}
func (msc *mockSnowthClient) ListInactiveNodes() []*gosnowth.SnowthNode {
	if msc.mockListInactiveNodes != nil {
		return msc.mockListInactiveNodes()
	}
	return nil
}

var gen = rand.New(rand.NewSource(2))

func getRandomNode(choices ...*gosnowth.SnowthNode) *gosnowth.SnowthNode {
	if len(choices) == 0 {
		return nil
	}
	choice := gen.Int() % len(choices)
	return choices[choice]
}

// MakeMetricList - given a prometheus MetricFamily pointer, an
// accountID, checkName and check UUID, generate the Flatbuffer
// serialized message that is to be sent to the IRONdb api.
// This will result in a []byte and error, where the []byte is
// the corresponding MetricList serialized
func MakeMetricList(promMetricFamily *dto.MetricFamily,
	accountID int32, checkName string, checkUUID uuid.UUID) ([]byte, error) {
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
		mOffset, err := MakeMetric(b, metric, *promMetricFamily.Type, accountID, checkName, checkUUID)
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
	b.FinishWithFileIdentifier(metricListOffset, []byte("CIML"))
	// return the finished serialized bytes
	return b.FinishedBytes(), nil
}

// MakeMetric - serialize a prometheus Metric as a flatbuffer resulting
// in the offset on the builder for the Metric
func MakeMetric(b *flatbuffers.Builder, promMetric *dto.Metric, metricType dto.MetricType,
	accountID int32, checkName string, checkUUID uuid.UUID) (flatbuffers.UOffsetT, error) {

	// prometheus metric types are as follows:
	// MetricType_COUNTER   MetricType = 0 -> NNT
	// MetricType_GAUGE     MetricType = 1 -> NNT
	// MetricType_SUMMARY   MetricType = 2 -> histogram
	// MetricType_UNTYPED   MetricType = 3 -> NNT/text?
	// MetricType_HISTOGRAM MetricType = 4 -> histogram

	var (
		// apply the checkName and UUID to the metric
		checkNameOffset = b.CreateString(checkName)
		checkUUIDOffset = b.CreateString(checkUUID.String())
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

	// TODO: if metric type is counter/gauge do the below,
	// if histogram/summary we need to use those union types.

	// create the metric value value
	circfb.DoubleValueStart(b)
	circfb.DoubleValueAddValue(b, promMetric.GetUntyped().GetValue())
	valueValue := circfb.DoubleValueEnd(b)

	// create the metric value
	circfb.MetricValueStart(b)
	// add timestamp to metric value
	var timestamp = uint64(promMetric.GetTimestampMs())
	if promMetric.GetTimestampMs() == 0 {
		// not here, we should add a timestamp
		timestamp = uint64(time.Now().UnixNano() / int64(time.Millisecond))
	}
	circfb.MetricValueAddTimestamp(b, timestamp)
	// add name to metric value
	circfb.MetricValueAddName(b, checkNameOffset)
	circfb.MetricValueAddStreamTags(b, streamTagVec)
	// this is the value of the value...
	circfb.MetricValueAddValueType(b, circfb.MetricValueUnionDoubleValue)
	circfb.MetricValueAddValue(b, valueValue)

	value := circfb.MetricValueEnd(b)
	// start a metric
	circfb.MetricStart(b)
	// add the timestamp
	circfb.MetricAddTimestamp(b, uint64(promMetric.GetTimestampMs()))
	// add the check name
	circfb.MetricAddCheckName(b, checkNameOffset)
	// add the check uuid
	circfb.MetricAddCheckUuid(b, checkUUIDOffset)
	// add the account ID to the Metric
	circfb.MetricAddAccountId(b, accountID)
	circfb.MetricAddValue(b, value)

	// alignment...
	b.Prep(1, flatbuffers.SizeInt32)
	b.PlaceInt32(0)
	metric := circfb.MetricEnd(b)
	// return the offset of the metric to the caller
	return metric, nil
}
