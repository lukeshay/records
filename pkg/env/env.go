package env

import "os"

func DefaultEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func RequireEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		panic("required environment variable not found: " + key)
	}
	return value
}
