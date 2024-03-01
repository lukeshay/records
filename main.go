package main

import (
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gofiber/template/html/v2"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/lukeshay/records/pkg/config"
	"github.com/lukeshay/records/pkg/database"
	"github.com/lukeshay/records/pkg/env"
	"github.com/lukeshay/records/pkg/log"
	"github.com/lukeshay/records/pkg/middleware"
	"github.com/lukeshay/records/pkg/routers"
	fibertrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gofiber/fiber.v2"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

//go:embed public/* templates/*
var assets embed.FS

func createTmplData(data map[string]any) map[string]any {
	data["Version"] = config.Version
	data["Environment"] = config.Environment
	data["DatadogClientToken"] = config.DatadogClientToken

	return data
}

func main() {
	ddAgent, ok := os.LookupEnv("DD_AGENT_HOST")
	if ok && config.Environment != "local" {
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
	database.InitDb()

	engine := html.NewFileSystem(http.FS(assets), ".html")
	engine.AddFunc("dict", func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	})
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Use(recover.New())
	app.Use(fibertrace.Middleware(fibertrace.WithServiceName("records")))
	app.Use(logger.New())
	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: config.CookieKey,
	}))
	app.Use("/public", filesystem.New(filesystem.Config{
		Root:       http.FS(assets),
		PathPrefix: "public",
		Browse:     config.Environment == "local",
	}))
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "pong",
		})
	})
	app.Get("/metadata", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"version":     config.Version,
			"environment": config.Environment,
		})
	})
	auth := app.Group("/auth")
	{
		auth.Use(middleware.UnauthRequired())
		auth.Get("/signin", routers.GetAuthSignIn)
		auth.Post("/signin", routers.PostAuthSignIn)
		auth.Get("/signup", routers.GetAuthSignUp)
		auth.Post("/signup", routers.PostAuthSignUp)
	}

	records := app.Group("/records")
	{
		records.Use(middleware.AuthRequired())
		records.Get("/", routers.GetRecords)
	}

	app.Listen(fmt.Sprintf(":%s", env.DefaultEnv("PORT", "8080")))
}
