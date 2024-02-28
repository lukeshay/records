package config

import "github.com/lukeshay/records/pkg/env"

type Config string

var (
	Version            = "undefined"
	Environment        = env.DefaultEnv("ENVIRONMENT", "local")
	DatadogClientToken = env.DefaultEnv("DATADOG_CLIENT_TOKEN", "")
	DatabaseURL        = env.RequireEnv("DATABASE_URL")
	DatabaseToken      = env.RequireEnv("DATABASE_TOKEN")
	CookieKey          = env.RequireEnv("COOKIE_KEY")
)
