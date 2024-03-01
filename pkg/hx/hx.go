package hx

import "github.com/gofiber/fiber/v2"

func AddRedirect(c *fiber.Ctx, path string) {
	c.Append("HX-Redirect", path)
}

func SendRedirect(c *fiber.Ctx, path string) error {
	AddRedirect(c, path)

	return c.Send([]byte{})
}
