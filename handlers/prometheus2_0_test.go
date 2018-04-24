package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
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

	url := fmt.Sprintf("/prometheus/2.0/write/42/checkname/%s", uuid.NewV4().String())
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

	url := fmt.Sprintf("/prometheus/2.0/read/42/checkname/%s", uuid.NewV4().String())
	r, _ := http.NewRequest("GET", url, bytes.NewBufferString(promUntypedMetric))
	w := httptest.NewRecorder()

	e.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("failure in write request: %s %d\n", w.Body.String(), w.Code)
	}

}
