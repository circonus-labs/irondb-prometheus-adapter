package handlers

import (
	"net/http"

	"github.com/circonus-labs/irondb-prometheus-adapter/renderings"
	"github.com/labstack/echo"
)

// HealthCheck - this is the health check handler, which informs about the
// health of the service.  Also includes important build information
func HealthCheck(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, renderings.HealthCheckResponse{
		Message:   "Success",
		CommitID:  ctx.Get("commitID").(string),
		BuildTime: ctx.Get("buildTime").(string),
	})
}
