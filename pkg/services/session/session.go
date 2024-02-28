package sessionservice

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lukeshay/records/pkg/database"
)

type key string

const (
	sessionCookieKey key = "session"
)

func SetSessionIDCookie(c *fiber.Ctx, session *database.Session) {
	c.Cookie(&fiber.Cookie{
		Name:  string(sessionCookieKey),
		Value: string(session.ID),
	})
}

func ClearSessionIDCookie(c *fiber.Ctx) {
	c.ClearCookie(string(sessionCookieKey))
}

func GetSessionIDCookie(c *fiber.Ctx) (int, error) {
	cookie := c.Cookies(string(sessionCookieKey))

	id, err := strconv.Atoi(cookie)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func GetSessionFromCookie(c *fiber.Ctx) (*database.Session, error) {
	id, err := GetSessionIDCookie(c)
	if err != nil {
		return nil, err
	}

	return database.SelectSessionByID(c.Context(), id)
}

func GetSessionFromContext(c *fiber.Ctx) (*database.Session, error) {
	session, err := c.Locals(string(sessionCookieKey)).(*database.Session)
	if !err {
		return nil, fmt.Errorf("session not found in context")
	}
	return session, nil
}

func SetSessionInContext(c *fiber.Ctx, session *database.Session) {
	c.Locals(string(sessionCookieKey), session)
}

func IsSessionValid(session *database.Session) bool {
	return session.ID > 0 && session.ExpiresAt.After(time.Now())
}

func GetValidSessionFromCookie(c *fiber.Ctx) (*database.Session, error) {
	session, err := GetSessionFromCookie(c)
	if err != nil {
		return nil, err
	}

	if !IsSessionValid(session) {
		ClearSessionIDCookie(c)
		return nil, fmt.Errorf("session has expired")
	}

	return session, nil
}

func GetSession(c *fiber.Ctx) (*database.Session, error) {
	session, err := GetSessionFromContext(c)
	if err == nil && session != nil {
		return session, nil
	}

	session, err = GetValidSessionFromCookie(c)
	if err != nil {
		return nil, err
	}

	SetSessionInContext(c, session)

	return session, nil
}
