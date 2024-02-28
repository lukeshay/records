package routers

import (
	"github.com/lukeshay/records/pkg/config"
)

func bindPage(data ...map[string]interface{}) map[string]interface{} {
	data = append(data, map[string]interface{}{
		"Version":            config.Version,
		"Environment":        config.Environment,
		"DatadogClientToken": config.DatadogClientToken,
	})
	result := make(map[string]interface{})
	for _, m := range data {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}
