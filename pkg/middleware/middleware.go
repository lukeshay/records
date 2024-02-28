package middleware

import (
	"github.com/gofiber/fiber/v2"
	sessionservice "github.com/lukeshay/records/pkg/services/session"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, err := sessionservice.GetSession(c)
		if err != nil {
			return c.Redirect("/auth/signin")
		}

		return c.Next()
	}
}

func UnauthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, err := sessionservice.GetSession(c)
		if err == nil {
			return c.Redirect("/")
		}
		return c.Next()
	}
}
