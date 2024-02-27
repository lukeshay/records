package config

import "github.com/lukeshay/records/pkg/env"

type Config string

var (
	Environment = env.DefaultEnv("ENVIRONMENT", "")
	Version     = "undefined"
)
