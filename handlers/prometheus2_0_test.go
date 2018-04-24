package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	circfb "github.com/circonus/promadapter/flatbuffer/circonus"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/labstack/echo"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	uuid "github.com/satori/go.uuid"
)

const (
	promUntypedMetric = `http_requests_total{method="post",code="200"} 1027 1395066363000
`
)

func TestPrometheusWrite2_0(t *testing.T) {
	// setup echo bits
	e := echo.New()
	e.POST("/prometheus/2.0/write/:account/:check_uuid/:check_name", PrometheusWrite2_0)

	url := fmt.Sprintf("/prometheus/2.0/write/42/checkname/%s", uuid.Must(uuid.NewV4()))
	r, _ := http.NewRequest("POST", url, bytes.NewBufferString(promUntypedMetric))
	w := httptest.NewRecorder()

	e.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("failure in write request: %s %d\n", w.Body.String(), w.Code)
	}
}

func TestPrometheusRead2_0(t *testing.T) {
	// setup echo bits
	e := echo.New()
	e.GET("/prometheus/2.0/read/:account/:check_uuid/:check_name", PrometheusWrite2_0)

	url := fmt.Sprintf("/prometheus/2.0/read/42/checkname/%s", uuid.Must(uuid.NewV4()))
	r, _ := http.NewRequest("GET", url, bytes.NewBufferString(promUntypedMetric))
	w := httptest.NewRecorder()

	e.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("failure in write request: %s %d\n", w.Body.String(), w.Code)
	}

}

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
	checkUUID := uuid.Must(uuid.NewV4()).String()
	metricOffset, err := MakeMetric(b, metricFamily.Metric[0], "42", "check_name", checkUUID)

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
	checkUUID := uuid.Must(uuid.NewV4()).String()
	data, err = MakeMetricList(metricFamily, "42", "check_name", checkUUID)

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
	}
}
