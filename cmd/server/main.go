// Package main - main entrypoint for the prometheus adapter application
package main

import (
	"flag"

	"github.com/circonus-labs/gosnowth"
	"github.com/circonus-labs/irondb-prometheus-adapter/handlers"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	uuid "github.com/satori/go.uuid"
)

var (
	// the application log level command line flag `-log`
	appLogLevel logLevel
	// the application listen address command line flag `-addr`
	addr string
	// the snowth address command line flags `-snowth`
	snowths snowthAddrFlag
	// commitID is populated at compile time with -ldflags '-X ...
	commitID string
	// buildTime is populated at compile time with -ldflags '-X ...
	buildTime string
)

// init - main init function which is used to parse the command line flags
func init() {
	flag.StringVar(&addr, "addr", ":8080", "Address for adapter to listen")
	flag.Var(&appLogLevel, "log", "Log level for adapter")
	flag.Var(&snowths, "snowth", "Snowth node to bootstrap")
	flag.Parse()

	log.Printf("addr flag: %s", addr)
	log.Printf("log flag: %s", appLogLevel)
	log.Printf("snowth flag: %+v", snowths)
}

// main - main entrypoint for irondb-prometheus-adapter application
func main() {
	// startup our gosnowth client
	snowthClient, err := gosnowth.NewSnowthClient(false, snowths...)
	if err != nil {
		log.Fatalf("failed to start snowth client: %s", err.Error())
	}

	e := echo.New()
	e.Logger.SetLevel(log.Lvl(appLogLevel))

	// Pre middleware which is applied to all routes; we are using this
	// to set the initial context with commitID/buildTime and a requestID
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// add the snowth client to the base context
			ctx.Set("snowthClient", snowthClient)
			// set the commit id of the build, so we can use that in our handlers
			ctx.Set("commitID", commitID)
			ctx.Set("buildTime", buildTime)
			ctx.Set("requestID", uuid.NewV4().String())
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
