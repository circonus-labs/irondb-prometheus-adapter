package handlers

import (
	"net/http"

	"github.com/circonus-labs/irondb-prometheus-adapter/renderings"
	"github.com/labstack/echo"
)

func HealthCheck(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, renderings.HealthCheckResponse{
		Message:   "Success",
		CommitID:  ctx.Get("commitID").(string),
		BuildTime: ctx.Get("buildTime").(string),
	})
}
