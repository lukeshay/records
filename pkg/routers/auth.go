package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lukeshay/records/pkg/database"
	sessionservice "github.com/lukeshay/records/pkg/services/session"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func GetAuthSignIn(c *fiber.Ctx) error {
	return c.Render("auth_signup", bindPage(fiber.Map{"IsSignUp": false}))
}

func PostAuthSignIn(c *fiber.Ctx) error {
	// ...
	return nil
}

func GetAuthSignUp(c *fiber.Ctx) error {
	return c.Render("auth_signup", bindPage(fiber.Map{"IsSignUp": true}))
}

type signUpData struct {
	Email          string `form:"email"`
	Password       string `form:"password"`
	RepeatPassword string `form:"repeatPassword"`
}

func (d *signUpData) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.Email, validation.Required, is.Email),
		validation.Field(&d.Password, validation.Required, validation.Length(12, 72)),
		validation.Field(
			&d.RepeatPassword,
			validation.Required,
			validation.NewStringRule(func(str string) bool {
				return str == d.Password
			}, "passwords do not match"),
		),
	)
}

func PostAuthSignUp(c *fiber.Ctx) error {
	body := &signUpData{}
	if err := c.BodyParser(body); err != nil {
		return c.Render("form_error", fiber.Map{
			"Error": err.Error(),
		})
	}

	if err := body.Validate(); err != nil {
		return c.Render("form_error", fiber.Map{
			"Error": err.Error(),
		})
	}

	existingUser, err := database.SelectUserByEmail(c.Context(), body.Email)
	if err != nil {
		return c.Render("form_error", fiber.Map{
			"Error": err.Error(),
		})
	}

	if existingUser != nil {
		return c.Render("form_error", fiber.Map{
			"Error": "user already exists",
		})
	}

	_, session, err := database.InsertUserAndSession(c.Context(), database.User{
		Email:          body.Email,
		HashedPassword: body.Password,
	})
	if err != nil {
		return c.Render("form_error", fiber.Map{
			"Error": err.Error(),
		})
	}

	sessionservice.SetSessionIDCookie(c, session)

	return c.Redirect("/")
}
