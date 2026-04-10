//go:build tools

// Package tools is used to track tool dependencies via go modules.
// These blank imports ensure all required packages remain in go.mod.
package tools

import (
	_ "github.com/go-chi/chi/v5"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "golang.org/x/crypto/bcrypt"
)
