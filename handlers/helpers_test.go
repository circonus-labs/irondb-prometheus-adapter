package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	circfb "github.com/circonus-labs/irondb-prometheus-adapter/flatbuffer/metrics"
	flatbuffers "github.com/google/flatbuffers/go"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	uuid "github.com/satori/go.uuid"
)

func TestMakeMetric(t *testing.T) {
	var (
		b            = flatbuffers.NewBuilder(0)
		dec          = expfmt.NewDecoder(bytes.NewBufferString(promUntypedMetric), "plain/text")
		metricFamily = new(dto.MetricFamily)
		err          error
	)

	// decode the metrics into the metric family
	err = dec.Decode(metricFamily)
	if err != nil {
		t.Errorf("failed to decode prometheus write message: %s", err.Error())
	}
	checkUUID := uuid.NewV4().String()
	metricOffset, err := MakeMetric(b, metricFamily.Metric[0], *metricFamily.Type, "42", "check_name", checkUUID)

	b.Finish(metricOffset)

	fbData := b.FinishedBytes()
	// now decode the flatbuffer and see if it looks right
	checkMetric := circfb.GetRootAsMetric(fbData, 0)

	if !bytes.Equal(checkMetric.CheckName(), []byte("check_name")) {
		t.Error("invalid check name")
	}
	if !bytes.Equal(checkMetric.CheckUuid(), []byte(checkUUID)) {
		t.Error("invalid check uuid")
	}
	if checkMetric.AccountId() != 42 {
		t.Error("invalid account id")
	}
	if checkMetric.Timestamp() != uint64(metricFamily.Metric[0].GetTimestampMs()) {
		t.Error("invalid account id")
	}
	checkValue := new(circfb.MetricValue)
	if value := checkMetric.Value(checkValue); value != nil {
		for i := 0; i < value.StreamTagsLength(); i++ {
			checkTags := value.StreamTags(i)
			if !strings.Contains(string(checkTags), ":") {
				t.Error("invalid stream tag format")
			}
		}
	}
}

func TestMakeMetricList(t *testing.T) {
	var (
		dec          = expfmt.NewDecoder(bytes.NewBufferString(promUntypedMetric), "plain/text")
		metricFamily = new(dto.MetricFamily)
		err          error
		data         []byte
	)

	// decode the metrics into the metric family
	err = dec.Decode(metricFamily)
	if err != nil {
		t.Errorf("failed to decode prometheus write message: %s", err.Error())
	}
	checkUUID := uuid.NewV4().String()
	data, err = MakeMetricList(metricFamily, "42", "check_name", checkUUID)

	fmt.Printf("bytes: %x\n", data)
	fmt.Printf("bytes: %s\n", hex.Dump(data))
	fmt.Printf("bytes: %s\n", base64.StdEncoding.EncodeToString(data))

	// now decode the flatbuffer and see if it looks right
	checkMetricList := circfb.GetRootAsMetricList(data, 0)

	if checkMetricList.MetricsLength() != 1 {
		t.Error("should only have one metric for this test")
	}

	checkMetric := new(circfb.Metric)
	if checkMetricList.Metrics(checkMetric, 0) {
		if !bytes.Equal(checkMetric.CheckName(), []byte("check_name")) {
			t.Error("invalid check name")
		}
		if !bytes.Equal(checkMetric.CheckUuid(), []byte(checkUUID)) {
			t.Error("invalid check uuid")
		}
		if checkMetric.AccountId() != 42 {
			t.Error("invalid account id")
		}
		if checkMetric.Timestamp() != uint64(metricFamily.Metric[0].GetTimestampMs()) {
			t.Error("invalid account id")
		}
		checkValue := new(circfb.MetricValue)
		checkMetric.Value(checkValue)

		unionTable := new(flatbuffers.Table)
		if checkValue.Value(unionTable) {
			unionType := checkValue.ValueType()
			if unionType == circfb.MetricValueUnionDoubleValue {
				checkValueValue := new(circfb.DoubleValue)
				checkValueValue.Init(unionTable.Bytes, unionTable.Pos)
				if checkValueValue.Value() != 1027 {
					t.Errorf("Value is not correct in flatbuffer %f\n", checkValueValue.Value())
				}
			}
		}
	}
}
