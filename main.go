package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/gin-gonic/gin"
	"github.com/lukeshay/records/pkg/config"
	"github.com/lukeshay/records/pkg/env"
	"github.com/lukeshay/records/pkg/log"
	gintrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func main() {
	ddAgent, ok := os.LookupEnv("DD_AGENT_HOST")
	if ok {
		tracer.Start(
			tracer.WithEnv(config.Environment),
			tracer.WithService("deployer"),
			tracer.WithServiceVersion(config.Version),
			tracer.WithLogStartup(false),
			tracer.WithDebugMode(false),
		)
		profiler.Start(
			profiler.WithEnv(config.Environment),
			profiler.WithService("deployer"),
			profiler.WithVersion(config.Version),
			profiler.WithProfileTypes(
				profiler.CPUProfile,
				profiler.HeapProfile,
			),
			profiler.WithLogStartup(false),
		)
		statsd.New(fmt.Sprintf("http://%s:8125", ddAgent))
	}

	defer tracer.Stop()

	log.InitLogger()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gintrace.Middleware("records"))
	r.Use(func(c *gin.Context) {
		args := []any{"method", c.Request.Method, "path", c.Request.URL.Path, "user_agent", c.Request.Header.Get("user-agent")}

		start := time.Now()

		c.Next()

		latency := time.Since(start)

		args = append(args, "latency", latency.Nanoseconds())

		if c.Errors.String() != "" {
			args = append(args, "error", c.Errors.String())
		}

		slog.DebugContext(c.Request.Context(), "Request", args...)
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run(fmt.Sprintf(":%s", env.DefaultEnv("PORT", "8080")))
}
