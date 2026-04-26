package db

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

// Player is one cached (puzzle leaderboard, nickname) → agentId mapping.
type Player struct {
	LeaderboardID string
	Nickname      string
	AgentID       int64
	FetchedAt     time.Time
}

// PlayerRepo persists Player rows.
type PlayerRepo struct {
	db *sql.DB
}

// Find returns the cached Player for the given leaderboard + nickname, or
// (nil, nil) if absent. Nickname matching is case-insensitive.
func (r *PlayerRepo) Find(leaderboardID, nickname string) (*Player, error) {
	var (
		p  Player
		ts int64
	)
	err := r.db.QueryRow(
		`SELECT leaderboard_id, nickname, agent_id, fetched_at
		 FROM players WHERE leaderboard_id = ? AND nickname = ?`,
		leaderboardID, strings.ToLower(nickname),
	).Scan(&p.LeaderboardID, &p.Nickname, &p.AgentID, &ts)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.FetchedAt = time.Unix(ts, 0)
	return &p, nil
}

// Save upserts a player mapping. Nickname is stored lower-cased for stable
// case-insensitive lookups.
func (r *PlayerRepo) Save(leaderboardID, nickname string, agentID int64) error {
	_, err := r.db.Exec(
		`INSERT INTO players(leaderboard_id, nickname, agent_id, fetched_at) VALUES(?, ?, ?, ?)
		 ON CONFLICT(leaderboard_id, nickname) DO UPDATE SET
		    agent_id   = excluded.agent_id,
		    fetched_at = excluded.fetched_at`,
		leaderboardID, strings.ToLower(nickname), agentID, time.Now().Unix(),
	)
	return err
}
