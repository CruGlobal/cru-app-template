// Package db connects to Cloud SQL Postgres. Nothing connects until the app
// calls Open, and the pool it returns opens connections only as they are
// needed, so an app with no database is unaffected.
//
// Modes, chosen from environment variables:
//
//	password: DATABASE_URL or DATABASE_PASSWORD is set. A plain pgx pool.
//	iam:      DATABASE_HOST, DATABASE_NAME and DATABASE_USER are set. Connects
//	          to the private IP and uses a Google login token as the password.
//	off:      anything else. Open returns an error.
package db

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2/google"
)

// Mode is how Open connects.
type Mode string

const (
	ModeOff      Mode = "off"
	ModePassword Mode = "password"
	ModeIAM      Mode = "iam"
)

var iamVars = []string{"DATABASE_HOST", "DATABASE_NAME", "DATABASE_USER"}

const loginScope = "https://www.googleapis.com/auth/sqlservice.login"

// ModeFrom picks the mode from an environment lookup such as os.Getenv.
func ModeFrom(getenv func(string) string) Mode {
	if getenv("DATABASE_URL") != "" || getenv("DATABASE_PASSWORD") != "" {
		return ModePassword
	}
	for _, name := range iamVars {
		if getenv(name) == "" {
			return ModeOff
		}
	}
	return ModeIAM
}

// Open returns a connection pool for the configured database. Calling it at
// startup is fine, because it connects lazily. The caller owns the pool and
// closes it on shutdown.
func Open(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := config(os.Getenv)
	if err != nil {
		return nil, err
	}

	if ModeFrom(os.Getenv) == ModeIAM {
		// Background, not ctx: the token source lives as long as the pool, and
		// a canceled ctx would break every later renewal.
		tokens, err := google.DefaultTokenSource(context.Background(), loginScope)
		if err != nil {
			return nil, fmt.Errorf("find Google credentials for the database: %w", err)
		}
		// The token source caches the token and renews it near expiry. pgx calls
		// this for each new connection, so a connection opened after a long idle
		// gap still gets a live token.
		cfg.BeforeConnect = func(_ context.Context, conn *pgx.ConnConfig) error {
			token, err := tokens.Token()
			if err != nil {
				return fmt.Errorf("get a Google login token for the database: %w", err)
			}
			conn.Password = token.AccessToken
			return nil
		}
	}

	return pgxpool.NewWithConfig(ctx, cfg)
}

// config builds the pool settings for the current mode.
func config(getenv func(string) string) (*pgxpool.Config, error) {
	switch ModeFrom(getenv) {
	case ModePassword:
		if dsn := getenv("DATABASE_URL"); dsn != "" {
			return pgxpool.ParseConfig(dsn)
		}
		// An empty host leaves pgx to its own default, PGHOST or else a local
		// server, as the other stacks' helpers do.
		u := url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(getenv("DATABASE_USER"), getenv("DATABASE_PASSWORD")),
			Host:   getenv("DATABASE_HOST"),
			Path:   "/" + getenv("DATABASE_NAME"),
		}
		return pgxpool.ParseConfig(u.String())

	case ModeIAM:
		// sslmode=require encrypts without checking the server certificate,
		// which a per-instance CA signs.
		u := url.URL{
			Scheme:   "postgres",
			User:     url.User(getenv("DATABASE_USER")),
			Host:     net.JoinHostPort(getenv("DATABASE_HOST"), "5432"),
			Path:     "/" + getenv("DATABASE_NAME"),
			RawQuery: "sslmode=require",
		}
		return pgxpool.ParseConfig(u.String())

	default:
		var missing []string
		for _, name := range iamVars {
			if getenv(name) == "" {
				missing = append(missing, name)
			}
		}
		return nil, errors.New("database is not configured: set DATABASE_URL or DATABASE_PASSWORD, or set " +
			strings.Join(missing, ", "))
	}
}
