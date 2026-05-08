package db

import (
	"database/sql"
	"errors"
	"time"
)

// Puzzle is one cached puzzle slug mapping. PrettyID is the slug found in
// CodinGame URLs (e.g. "<season>-<year>-<game>"); LeaderboardID is the
// internal slug used by the puzzle leaderboard endpoints
// (e.g. "<season>-<year>-<sponsor>").
type Puzzle struct {
	PrettyID      string
	LeaderboardID string
	FetchedAt     time.Time
}

// PuzzleRepo persists Puzzle rows.
type PuzzleRepo struct {
	db *sql.DB
}

// Find returns the cached Puzzle for prettyID, or (nil, nil) if absent.
func (r *PuzzleRepo) Find(prettyID string) (*Puzzle, error) {
	var (
		p  Puzzle
		ts int64
	)
	err := r.db.QueryRow(
		`SELECT pretty_id, leaderboard_id, fetched_at FROM puzzles WHERE pretty_id = ?`,
		prettyID,
	).Scan(&p.PrettyID, &p.LeaderboardID, &ts)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.FetchedAt = time.Unix(ts, 0)
	return &p, nil
}

// Save upserts a puzzle slug mapping.
func (r *PuzzleRepo) Save(prettyID, leaderboardID string) error {
	_, err := r.db.Exec(
		`INSERT INTO puzzles(pretty_id, leaderboard_id, fetched_at) VALUES(?, ?, ?)
		 ON CONFLICT(pretty_id) DO UPDATE SET
		    leaderboard_id = excluded.leaderboard_id,
		    fetched_at     = excluded.fetched_at`,
		prettyID, leaderboardID, time.Now().Unix(),
	)
	return err
}
