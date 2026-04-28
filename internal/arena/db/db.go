// Package db is the arena's app-wide SQLite store. A single file in the
// current working directory holds cached lookups (puzzle slugs, player agent
// IDs) and any future arena state we don't want to re-fetch every run.
package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Path is the on-disk filename for the arena database. It lives in the
// current working directory so each project (or run-dir) has its own cache.
const Path = "db.sqlite3"

// DB wraps a *sql.DB and exposes typed repositories for each domain.
type DB struct {
	sql      *sql.DB
	Puzzles  *PuzzleRepo
	Players  *PlayerRepo
}

// Open opens (or creates) the SQLite database at path and applies the
// schema. Pass an empty path to use the default Path.
func Open(path string) (*DB, error) {
	if path == "" {
		path = Path
	}
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	if _, err := conn.Exec(schema); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}
	return &DB{
		sql:     conn,
		Puzzles: &PuzzleRepo{db: conn},
		Players: &PlayerRepo{db: conn},
	}, nil
}

// Close releases the underlying database handle.
func (d *DB) Close() error {
	if d == nil || d.sql == nil {
		return nil
	}
	return d.sql.Close()
}
