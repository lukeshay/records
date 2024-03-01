package sessionservice

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lukeshay/records/pkg/database"
	"golang.org/x/crypto/bcrypt"
)

type key string

const (
	sessionCookieKey key = "session"
)

func setSessionIDCookie(c *fiber.Ctx, session *database.Session) {
	slog.Debug("setSessionIDCookie")

	c.Cookie(&fiber.Cookie{
		Name:  string(sessionCookieKey),
		Value: strconv.FormatInt(session.ID, 10),
	})
}

func clearSessionIDCookie(c *fiber.Ctx) {
	slog.Debug("clearSessionIDCookie")

	c.ClearCookie(string(sessionCookieKey))
}

func getSessionIDCookie(c *fiber.Ctx) (int, error) {
	slog.Debug("getSessionIDCookie")
	cookie := c.Cookies(string(sessionCookieKey))

	slog.Debug("session id cookie", "cookie", cookie)

	if cookie == "" {
		return -1, fmt.Errorf("session id cookie not found")
	}

	id, err := strconv.Atoi(cookie)
	if err != nil {
		clearSessionIDCookie(c)
		return -1, err
	}

	return id, nil
}

func getSessionFromCookie(c *fiber.Ctx) (*database.Session, error) {
	slog.Debug("getSessionFromCookie")

	id, err := getSessionIDCookie(c)
	if err != nil {
		slog.Debug("error getting session id from cookie", "err", err.Error())
		return nil, err
	}

	slog.Debug("selecting session by id", "sessionID", id)

	return database.SelectSessionByID(c.Context(), id)
}

func getSessionFromContext(c *fiber.Ctx) (*database.Session, error) {
	slog.Debug("getSessionFromContext")

	session, found := c.Locals(string(sessionCookieKey)).(*database.Session)
	if !found || session == nil {
		slog.Debug("session not found in context")
		return nil, fmt.Errorf("session not found in context")
	}

	slog.Debug("session found in context", "sessionID", session.ID)

	return session, nil
}

func setSessionInContext(c *fiber.Ctx, session *database.Session) {
	c.Locals(string(sessionCookieKey), session)
}

func isSessionValid(session *database.Session) bool {
	return session.ID > 0 && session.ExpiresAt.After(time.Now())
}

func getValidSessionFromCookie(c *fiber.Ctx) (*database.Session, error) {
	slog.Debug("getValidSessionFromCookie")

	session, err := getSessionFromCookie(c)
	if err != nil || session == nil {
		slog.Debug("session not found in cookie")

		return session, err
	}

	slog.Debug("session found in cookie", "sessionID", session.ID)

	if !isSessionValid(session) {
		slog.Debug("session has expired", "sessionID", session.ID)
		clearSessionIDCookie(c)
		return nil, fmt.Errorf("session has expired")
	}

	return session, nil
}

func GetSession(c *fiber.Ctx) (*database.Session, error) {
	slog.Debug("GetSession")

	session, err := getSessionFromContext(c)
	if err == nil {
		return session, nil
	}

	session, err = getValidSessionFromCookie(c)
	if err != nil {
		return nil, err
	}

	setSessionInContext(c, session)

	return session, nil
}

func CreateSessionForUser(c *fiber.Ctx, user *database.User) (*database.Session, error) {
	session, err := database.InsertSession(c.Context(), database.Database, database.Session{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	})
	if err != nil {
		return nil, err
	}

	setSessionIDCookie(c, session)

	return session, nil
}

func CreateUserAndSession(c *fiber.Ctx, email, password string) (*database.User, *database.Session, error) {
	slog.Debug("hashing password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		slog.Error("error generating password", "err", err.Error())
		return nil, nil, err
	}

	slog.Debug("inserting user and session")
	newUser, session, err := database.InsertUserAndSession(c.Context(), database.User{
		Email:          email,
		HashedPassword: string(hashedPassword),
	})
	if err != nil {
		slog.Error("error inserting user and session", "err", err.Error())
		return nil, nil, err
	}

	setSessionIDCookie(c, session)

	return newUser, session, nil
}
