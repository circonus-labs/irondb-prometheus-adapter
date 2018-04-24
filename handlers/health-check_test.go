package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/circonus/promadapter/renderings"
	"github.com/labstack/echo"
)

func TestHealthCheck(t *testing.T) {
	// setup echo bits
	e := echo.New()
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// set the commit id of the build, so we can use that in our handlers
			ctx.Set("commitID", "commitID")
			ctx.Set("buildTime", "buildTime")
			return next(ctx)
		}
	})
	e.GET("/health-check", HealthCheck)

	r, _ := http.NewRequest("GET", "/health-check", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, r)

	hcr := new(renderings.HealthCheckResponse)
	json.Unmarshal(w.Body.Bytes(), hcr)

	if hcr.CommitID != "commitID" || hcr.BuildTime != "buildTime" {
		t.Errorf("invalid commit or build time")
	}
}
