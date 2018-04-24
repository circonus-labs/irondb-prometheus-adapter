// Package main - main entrypoint for the prometheus adapter application
package main

import (
	"flag"

	"github.com/circonus-labs/irondb-prometheus-adapter/handlers"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	uuid "github.com/satori/go.uuid"
)

var (
	appLogLevel logLevel
	addr        string
	snowths     snowthAddrFlag
	commitID    string
	buildTime   string
)

func init() {
	flag.StringVar(&addr, "addr", ":8080", "Address for adapter to listen")
	flag.Var(&appLogLevel, "log", "Log level for adapter")
	flag.Var(&snowths, "snowth", "Snowth node to bootstrap")
	flag.Parse()
}

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.Lvl(appLogLevel))

	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// set the commit id of the build, so we can use that in our handlers
			ctx.Set("commitID", commitID)
			ctx.Set("buildTime", buildTime)
			ctx.Set("requestID", uuid.Must(uuid.NewV4()))
			return next(ctx)
		}
	})

	// Routes
	e.POST("/prometheus/2.0/write/:account/:check_uuid/:check_name", handlers.PrometheusWrite2_0)
	e.GET("/prometheus/2.0/read/:account/:check_uuid/:check_name", handlers.PrometheusRead2_0)
	e.GET("/health-check", handlers.HealthCheck)

	// Start server
	e.Logger.Fatal(e.Start(addr))
}
