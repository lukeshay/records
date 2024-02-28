package routers

import "github.com/gofiber/fiber/v2"

func GetRecords(c *fiber.Ctx) error {
	return c.Render("index", bindPage())
}
