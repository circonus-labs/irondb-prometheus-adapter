// Package main - main entrypoint for the prometheus adapter application
package main

import (
	"flag"

	"github.com/circonus/promadapter/handlers"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

var (
	appLogLevel logLevel
	addr        string
	snowths     snowthAddrFlag
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

	// Routes
	e.POST("/prometheus/2.0/write/:account/:check_uuid/:check_name", handlers.PrometheusWrite2_0)
	e.GET("/prometheus/2.0/read/:account/:check_uuid/:check_name", handlers.PrometheusRead2_0)
	e.GET("/health-check", handlers.HealthCheck)

	// Start server
	e.Logger.Fatal(e.Start(addr))
}
