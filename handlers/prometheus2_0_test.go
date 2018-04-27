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
http_requests_total{method="post",code="200"} 1027 1395066363001
`
)

func TestPrometheusWrite2_0(t *testing.T) {
	// setup echo bits
	e := echo.New()

	// mock snowth client
	snowthClient := new(mockSnowthClient)
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// add the snowth client to the base context
			ctx.Set("snowthClient", snowthClient)
			return next(ctx)
		}
	})
	e.POST("/prometheus/2.0/write/:account/:check_uuid/:check_name", PrometheusWrite2_0)

	url := fmt.Sprintf("/prometheus/2.0/write/42/%s/check_name", uuid.NewV4().String())
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
	snowthClient := new(mockSnowthClient)
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// add the snowth client to the base context
			ctx.Set("snowthClient", snowthClient)
			return next(ctx)
		}
	})
	e.GET("/prometheus/2.0/read/:account/:check_uuid/:check_name", PrometheusRead2_0)

	url := fmt.Sprintf("/prometheus/2.0/read/42/checkname/%s", uuid.NewV4().String())
	r, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("failure in write request: %s %d\n", w.Body.String(), w.Code)
	}

}
