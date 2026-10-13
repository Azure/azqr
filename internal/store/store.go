// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package store provides persistence for azqr scan sessions. By default it is
// built on the pure-Go modernc.org/sqlite driver so the CLI and MCP server
// binaries keep building with CGO_ENABLED=0, matching the existing build
// matrix in the Makefile. Setting the AZQR_POSTGRESQL environment variable to
// a non-empty Postgres connection string switches the backend to Postgres
// (via the pure-Go jackc/pgx/v5 driver) for teams that want a shared,
// centrally-queryable store instead of a local SQLite file.
package store

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// dialect identifies which database backend a Store is using. It selects the
// SQL placeholder style (sqlx.Rebind), the goose migration dialect, and the
// connection pool sizing applied in Open.
type dialect string

const (
	dialectSQLite   dialect = "sqlite"
	dialectPostgres dialect = "postgres"
)

// postgresDSNEnvVar is the environment variable that, when set to a non-empty
// Postgres connection string, switches Store from its default SQLite file to
// Postgres. It takes precedence over the path/--store-path argument passed to
// Open, since a shared Postgres store is not addressed by filesystem path.
const postgresDSNEnvVar = "AZQR_POSTGRESQL"

// Store wraps a database holding every scan session azqr has recorded. On
// SQLite, writes must go through the single writer connection
// (db.SetMaxOpenConns(1)); reads may use additional connections against the
// same *sql.DB, since SQLite's WAL mode allows many concurrent readers. On
// Postgres, the server itself arbitrates concurrent writers, so a small fixed
// connection pool is used instead.
type Store struct {
	db      *sql.DB
	path    string
	dialect dialect
}

// DefaultPath returns the default central database path (~/.azqr/azqr.db),
// creating the parent directory if it does not already exist.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	dir := filepath.Join(home, ".azqr")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create store directory %q: %w", dir, err)
	}
	return filepath.Join(dir, "azqr.db"), nil
}

// Open opens (creating if necessary) the configured database, applies
// backend-appropriate connection settings, runs any pending migrations, and
// returns a ready-to-use Store.
//
// When the AZQR_POSTGRESQL environment variable is set to a non-empty
// Postgres connection string, Open connects to that Postgres database and
// path is ignored. Otherwise Open uses the SQLite file at path, resolving an
// empty path to DefaultPath().
func Open(path string) (*Store, error) {
	if dsn := os.Getenv(postgresDSNEnvVar); dsn != "" {
		return openPostgres(dsn)
	}
	return openSQLite(path)
}

// openSQLite opens (creating if necessary) the SQLite database at path,
// applies PRAGMA settings required for correct concurrent/batched operation,
// runs any pending migrations, and returns a ready-to-use Store.
//
// An empty path resolves to DefaultPath().
func openSQLite(path string) (*Store, error) {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return nil, err
		}
	}

	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database %q: %w", path, err)
	}

	// A single writer connection avoids SQLITE_BUSY errors from concurrent
	// writers; WAL mode still allows other processes/connections to read
	// concurrently. Every connection in the pool applies the same DSN
	// pragmas above, including foreign_keys, since SQLite enforces foreign
	// keys per connection rather than per database file.
	db.SetMaxOpenConns(1)

	if err := migrate(db, dialectSQLite); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate sqlite database %q: %w", path, err)
	}

	return &Store{db: db, path: path, dialect: dialectSQLite}, nil
}

// postgresMaxOpenConns bounds the connection pool used against a Postgres
// store. Unlike SQLite (one writer, serialized by the file itself), Postgres
// arbitrates concurrent writers server-side, so a small fixed pool is enough
// for azqr's single-process write pattern without needing a user-tunable
// setting.
const postgresMaxOpenConns = 5

// openPostgres connects to the Postgres database identified by dsn, applies
// a conservative fixed connection pool, runs any pending migrations, and
// returns a ready-to-use Store.
func openPostgres(dsn string) (*Store, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres database: %w", err)
	}

	db.SetMaxOpenConns(postgresMaxOpenConns)

	if err := migrate(db, dialectPostgres); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate postgres database: %w", err)
	}

	return &Store{db: db, path: "", dialect: dialectPostgres}, nil
}

// Path returns the filesystem path of the opened database. It is empty when
// the Store is backed by Postgres.
func (s *Store) Path() string {
	return s.path
}

// Close releases the underlying database connection.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// rebind translates a query written with SQLite/MySQL-style `?` positional
// placeholders into the placeholder syntax the Store's active dialect
// expects. On SQLite it returns query unchanged; on Postgres it rewrites `?`
// into `$1, $2, ...` via sqlx.Rebind. Every hand-written query executed
// against s.db (or a *sql.Tx derived from it) must be passed through rebind
// so the same query text works against either backend.
func (s *Store) rebind(query string) string {
	if s.dialect == dialectPostgres {
		return sqlx.Rebind(sqlx.DOLLAR, query)
	}
	return query
}

// migrate applies any pending embedded goose migrations for the given dialect.
func migrate(db *sql.DB, d dialect) error {
	goose.SetBaseFS(migrationFiles)
	goose.SetLogger(goose.NopLogger())

	gooseDialect := goose.DialectSQLite3
	if d == dialectPostgres {
		gooseDialect = goose.DialectPostgres
	}
	if err := goose.SetDialect(string(gooseDialect)); err != nil {
		return err
	}
	if err := goose.Up(db, "migrations"); err != nil && !errors.Is(err, goose.ErrNoNextVersion) {
		return err
	}
	return nil
}
