// Package db is the arena's app-wide SQLite store. A single file in the
// system temp directory holds cached lookups (puzzle slugs, player agent
// IDs) and any future arena state we don't want to re-fetch every run.
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Path is the on-disk path for the arena database. It lives in the system
// temp directory so the cache is shared across runs and projects.
var Path = filepath.Join(os.TempDir(), "cg-arena-db.sqlite3")

// DB wraps a *sql.DB and exposes typed repositories for each domain.
type DB struct {
	sql      *sql.DB
	Puzzles  *PuzzleRepo
	Players  *PlayerRepo
}

// Open opens (or creates) the SQLite database at path and applies the
// schema. Pass an empty path to use the default Path. If the on-disk
// schema version does not match SchemaVersion the file is removed and
// recreated.
func Open(path string) (*DB, error) {
	if path == "" {
		path = Path
	}
	conn, err := openSchema(path)
	if err != nil {
		return nil, err
	}

	var version string
	err = conn.QueryRow(`SELECT version FROM schema_version LIMIT 1`).Scan(&version)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = conn.Close()
		return nil, fmt.Errorf("read schema version: %w", err)
	}

	if version != SchemaVersion {
		_ = conn.Close()
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("remove stale db %s: %w", path, err)
		}
		conn, err = openSchema(path)
		if err != nil {
			return nil, err
		}
		if _, err := conn.Exec(`INSERT INTO schema_version(version) VALUES (?)`, SchemaVersion); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("write schema version: %w", err)
		}
	}

	return &DB{
		sql:     conn,
		Puzzles: &PuzzleRepo{db: conn},
		Players: &PlayerRepo{db: conn},
	}, nil
}

func openSchema(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	if _, err := conn.Exec(schema); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}
	return conn, nil
}

// Close releases the underlying database handle.
func (d *DB) Close() error {
	if d == nil || d.sql == nil {
		return nil
	}
	return d.sql.Close()
}
