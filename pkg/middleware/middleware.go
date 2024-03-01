package middleware

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/lukeshay/records/pkg/hx"
	sessionservice "github.com/lukeshay/records/pkg/services/session"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		slog.Debug("AuthRequired")

		_, err := sessionservice.GetSession(c)
		if err != nil {
			slog.Debug("error getting session", "err", err.Error())

			hx.AddRedirect(c, "/auth/signin")
			return c.Redirect("/auth/signin")
		}

		return c.Next()
	}
}

func UnauthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		slog.Debug("UnauthRequired")

		_, err := sessionservice.GetSession(c)
		if err == nil {
			hx.AddRedirect(c, "/records")
			return c.Redirect("/records")
		}
		return c.Next()
	}
}
