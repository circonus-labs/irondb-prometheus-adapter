// Package main - main entrypoint for the prometheus adapter application
package main

import (
	"flag"
	"time"

	"github.com/circonus-labs/gosnowth"
	"github.com/circonus-labs/irondb-prometheus-adapter/handlers"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	uuid "github.com/satori/go.uuid"
)

var (
	// the application log level command line flag `-log`
	appLogLevel = logLevel(log.ERROR)
	// the application listen address command line flag `-addr`
	addr string
	// the snowth address command line flags `-snowth`
	snowths snowthAddrFlag
	// commitID is populated at compile time with -ldflags '-X ...
	commitID string
	// buildTime is populated at compile time with -ldflags '-X ...
	buildTime string
	// read off
	readOff bool
	// write off
	writeOff bool
)

// init - main init function which is used to parse the command line flags
func init() {
	flag.StringVar(&addr, "addr", ":8080", "Address for adapter to listen")
	flag.Var(&appLogLevel, "log", "Log level for adapter")
	flag.Var(&snowths, "snowth", "Snowth node to bootstrap")
	flag.BoolVar(&readOff, "readOff", false, "Turn read endpoint off")
	flag.BoolVar(&writeOff, "writeOff", false, "Turn write endpoint off")
	flag.Parse()

	log.Printf("addr flag: %s", addr)
	log.Printf("log flag: %s", appLogLevel)
	log.Printf("snowth flag: %+v", snowths)
	log.Printf("readOff flag: %+v", readOff)
	log.Printf("writeOff flag: %+v", writeOff)
}

// main - main entrypoint for irondb-prometheus-adapter application
func main() {
	// startup our gosnowth client
	// if no snowth nodes are available, sleep and loop until they are (1000 times max)
	// avoids the race condition where adapter is started w/ no snowth nodes available
	var snowthClient *gosnowth.SnowthClient
	var err error
	maxRetries := 1000
	for i := 0; i < maxRetries; i++ {
		if snowthClient, err = gosnowth.NewSnowthClient(false, snowths...); err == nil {
			break
		}

		log.Printf("No IRONdb nodes available on attempt %d, trying again in 10 seconds: %s", i+1, err.Error())
		time.Sleep(10 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to make connection to IRONdb after %d attempts, last error: %v", maxRetries+1, err.Error())
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
	if !writeOff {
		e.POST("/prometheus/2.0/write/:account/:check_uuid/:check_name", handlers.PrometheusWrite2_0)
	}
	if !readOff {
		e.POST("/prometheus/2.0/read/:account/:check_uuid/:check_name", handlers.PrometheusRead2_0)
	}
	e.GET("/health-check", handlers.HealthCheck)

	// Start server
	e.Logger.Fatal(e.Start(addr))
}
