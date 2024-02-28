package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lukeshay/records/pkg/config"
	"github.com/tursodatabase/libsql-client-go/libsql"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	sqlxtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/jmoiron/sqlx"
)

var Database *sqlx.DB

func InitDb() {
	url := fmt.Sprintf("%s?authToken=%s", config.DatabaseURL, config.DatabaseToken)

	sqltrace.Register("libsql", &libsql.Driver{}, sqltrace.WithServiceName("records"))

	var err error
	Database, err = sqlxtrace.Connect("libsql", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", url, err)
		os.Exit(1)
	}
	defer Database.Close()

	Database.Exec(`
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY,
      email TEXT NOT NULL,
      hashed_password TEXT NOT NULL
    )`,
	)
	Database.Exec(`
    CREATE TABLE IF NOT EXISTS sessions (
      id INTEGER PRIMARY KEY,
      user_id INTEGER NOT NULL,
      expires_at TIMESTAMP NOT NULL
    )`,
	)
}

type User struct {
	ID             int64  `json:"id" db:"id"`
	Email          string `json:"email" db:"email"`
	HashedPassword string `json:"-" db:"hashed_password"`
}

type Session struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}

func SelectUserByID(ctx context.Context, id int) (*User, error) {
	var user User
	err := Database.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", id)
	return &user, err
}

func SelectUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := Database.GetContext(ctx, &user, "SELECT * FROM users WHERE email = $1", email)
	return &user, err
}

func InsertUser(ctx context.Context, tx sqlx.ExecerContext, user User) (*User, error) {
	result, err := tx.ExecContext(ctx, "INSERT INTO users (email, hashed_password) VALUES ($1, $2) RETURNING *", user.Email, user.HashedPassword)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	user.ID = id

	return &user, nil
}

func SelectSessionByID(ctx context.Context, id int) (*Session, error) {
	var session Session
	err := Database.GetContext(ctx, &session, "SELECT * FROM sessions WHERE id = $1", id)
	return &session, err
}

func InsertSession(ctx context.Context, tx sqlx.ExecerContext, session Session) (*Session, error) {
	result, err := tx.ExecContext(ctx, "INSERT INTO sessions (user_id, expires_at) VALUES ($1, $2) RETURNING *", session.UserID, session.ExpiresAt)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	session.ID = id

	return &session, nil
}

func InsertUserAndSession(ctx context.Context, user User) (*User, *Session, error) {
	tx := Database.MustBeginTx(ctx, &sql.TxOptions{})

	insertedUser, err := InsertUser(ctx, tx, user)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	session, err := InsertSession(ctx, tx, Session{
		UserID:    insertedUser.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	})
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, nil, err
	}

	return insertedUser, session, nil
}
