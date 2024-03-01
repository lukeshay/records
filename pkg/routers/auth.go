package routers

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/lukeshay/records/pkg/database"
	"github.com/lukeshay/records/pkg/hx"
	sessionservice "github.com/lukeshay/records/pkg/services/session"
	"golang.org/x/crypto/bcrypt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func GetAuthSignIn(c *fiber.Ctx) error {
	return c.Render("auth_signup", bindPage(fiber.Map{"IsSignUp": false}))
}

type signInData struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func PostAuthSignIn(c *fiber.Ctx) error {
	slog.Debug("PostAuthSignIn")

	body := &signInData{}
	if err := c.BodyParser(body); err != nil {
		slog.Debug("error parsing body", "err", err.Error())
		return c.Render("form_error", fiber.Map{
			"Message": err.Error(),
		})
	}

	slog.Debug("selecting user by email", "email", body.Email)
	user, err := database.SelectUserByEmail(c.Context(), body.Email)
	if err != nil {
		return c.Render("form_error", fiber.Map{
			"Message": "invalid email or password",
		})
	}

	slog.Debug("comparing hashed password")
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(body.Password))
	if err != nil {
		return c.Render("form_error", fiber.Map{
			"Message": "invalid email or password",
		})
	}

	slog.Debug("creating session for user")
	_, err = sessionservice.CreateSessionForUser(c, user)
	if err != nil {
		slog.Error("error creating session for user", "err", err.Error())
		return c.Render("form_error", fiber.Map{
			"Message": "an unexpected error occurred. please try again later",
		})
	}

	return hx.SendRedirect(c, "/records")
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
	return validation.ValidateStruct(d,
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
	slog.Debug("PostAuthSignUp")

	body := &signUpData{}
	if err := c.BodyParser(body); err != nil {
		slog.Debug("error parsing body", "err", err.Error())
		return c.Render("form_error", fiber.Map{
			"Message": err.Error(),
		})
	}

	slog.Debug("validating body")
	if err := body.Validate(); err != nil {
		slog.Debug("error validating body", "err", err.Error())
		return c.Render("form_error", fiber.Map{
			"Message": err.Error(),
		})
	}

	slog.Debug("selecting user by email", "email", body.Email)
	_, err := database.SelectUserByEmail(c.Context(), body.Email)
	if err != nil {
		slog.Debug("error selecting user by email", "email", body.Email, "err", err.Error())

		if !database.IsNoRowsError(err) {
			slog.Error("error selecting user by email", "email", body.Email, "err", err.Error())
			return c.Render("form_error", fiber.Map{
				"Message": err.Error(),
			})
		}
	} else {
		slog.Debug("email is taken", "email", body.Email)
		return c.Render("form_error", fiber.Map{
			"Message": "email is taken",
		})
	}

	user, session, err := sessionservice.CreateUserAndSession(c, body.Email, body.Password)
	if err != nil {
		slog.Error("error creating user and session", "err", err.Error())

		return c.Render("form_error", fiber.Map{
			"Message": "an unexpected error occurred. please try again later",
		})
	}

	slog.Debug("created user and session", "userID", user.ID, "sessionID", session.ID)

	return hx.SendRedirect(c, "/records")
}
