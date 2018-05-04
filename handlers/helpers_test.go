package handlers

import (
	"bytes"
	"testing"
	"time"

	circfb "github.com/circonus-labs/irondb-prometheus-adapter/flatbuffer/metrics"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/prometheus/prometheus/prompb"
	uuid "github.com/satori/go.uuid"
)

func TestMakeMetric(t *testing.T) {
	var (
		req       prompb.WriteRequest
		checkUUID = uuid.NewV4()
		timestamp = int64(time.Now().Unix())
	)

	req.Timeseries = []*prompb.TimeSeries{
		&prompb.TimeSeries{
			Samples: []*prompb.Sample{
				&prompb.Sample{
					Timestamp: timestamp,
					Value:     42,
				},
			},
			Labels: []*prompb.Label{
				&prompb.Label{
					Name: "label-name", Value: "label-value",
				},
			},
		},
	}

	metricListBytes, err := MakeMetricList(
		req.GetTimeseries(), 42, "check_name", checkUUID)
	if err != nil {
		t.Error("error making metric list: ", err.Error())
	}

	// now decode the flatbuffer and see if it looks right
	checkMetricList := circfb.GetRootAsMetricList(metricListBytes, 0)

	if checkMetricList.MetricsLength() != 1 {
		t.Error("should only have one metric for this test")
	}

	checkMetric := new(circfb.Metric)
	if checkMetricList.Metrics(checkMetric, 0) {
		if !bytes.Equal(checkMetric.CheckName(), []byte("check_name")) {
			t.Error("invalid check name")
		}
		if cUUID := uuid.FromStringOrNil(string(checkMetric.CheckUuid())); !bytes.Equal(cUUID.Bytes(), checkUUID.Bytes()) {
			t.Error("invalid check uuid")
		}
		if checkMetric.AccountId() != 42 {
			t.Error("invalid account id")
		}
		if checkMetric.Timestamp() != uint64(timestamp) {
			t.Error("invalid timestamp")
		}
		checkValue := new(circfb.MetricValue)
		checkMetric.Value(checkValue)

		unionTable := new(flatbuffers.Table)
		if checkValue.Value(unionTable) {
			unionType := checkValue.ValueType()
			if unionType == circfb.MetricValueUnionDoubleValue {
				checkValueValue := new(circfb.DoubleValue)
				checkValueValue.Init(unionTable.Bytes, unionTable.Pos)
				if checkValueValue.Value() != 42 {
					t.Errorf("Value is not correct in flatbuffer %f\n", checkValueValue.Value())
				}
			}
		}
	}
}
