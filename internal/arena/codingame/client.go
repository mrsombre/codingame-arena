// Package codingame is a thin client for the public CodinGame.com endpoints
// the arena uses to download replays and resolve puzzle / player metadata.
// All Fetch* methods return the raw JSON body so callers can persist it
// untouched.
package codingame

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Endpoint URLs (POST, application/json).
const (
	PuzzleAPI      = "https://www.codingame.com/services/Puzzle/findProgressByPrettyId"
	LeaderboardAPI = "https://www.codingame.com/services/Leaderboards/getFilteredPuzzleLeaderboard"
	LastBattlesAPI = "https://www.codingame.com/services/gamesPlayersRanking/findLastBattlesByAgentId"
	GameResultAPI  = "https://www.codingame.com/services/gameResult/findInformationById"
)

// Client wraps an http.Client and exposes typed helpers for the CodinGame
// public services. The zero value uses http.DefaultClient.
type Client struct {
	HTTP *http.Client
}

// New returns a Client backed by http.DefaultClient.
func New() *Client { return &Client{} }

func (c *Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

// FetchReplay returns the raw replay JSON body for a CodinGame gameId.
func (c *Client) FetchReplay(gameID int64) ([]byte, error) {
	return c.post(GameResultAPI, fmt.Sprintf("[%d,null]", gameID))
}

// ResolvePuzzle maps a URL pretty-id (e.g. "winter-challenge-2026-snakebyte")
// to the puzzleLeaderboardId used by the leaderboard endpoints (e.g.
// "winter-challenge-2026-exotec").
func (c *Client) ResolvePuzzle(prettyID string) (string, error) {
	body, err := c.post(PuzzleAPI, fmt.Sprintf("[%q,null]", prettyID))
	if err != nil {
		return "", fmt.Errorf("resolve puzzle %q: %w", prettyID, err)
	}
	var resp struct {
		PuzzleLeaderboardID string `json:"puzzleLeaderboardId"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parse puzzle response: %w", err)
	}
	if resp.PuzzleLeaderboardID == "" {
		return "", fmt.Errorf("puzzle %q has no puzzleLeaderboardId", prettyID)
	}
	return resp.PuzzleLeaderboardID, nil
}

// AgentInfo bundles the leaderboard fields the arena reads for a player.
// Rank is the player's overall position; Division mirrors the API's
// `league.divisionIndex`. Score is the elo-like value the API returns.
type AgentInfo struct {
	AgentID  int64
	Rank     int
	Division int
	Score    float64
}

// FindAgent searches the leaderboard for a nickname and returns the matching
// AgentInfo for the given puzzle leaderboard slug. Matches case-insensitively.
func (c *Client) FindAgent(apiSlug, nickname string) (AgentInfo, error) {
	payload, err := json.Marshal([]any{
		apiSlug,
		nil,
		"global",
		map[string]any{"active": true, "column": "KEYWORD", "filter": nickname},
	})
	if err != nil {
		return AgentInfo{}, err
	}
	body, err := c.post(LeaderboardAPI, string(payload))
	if err != nil {
		return AgentInfo{}, fmt.Errorf("search leaderboard: %w", err)
	}
	var resp struct {
		Users []leaderboardUser `json:"users"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return AgentInfo{}, fmt.Errorf("parse leaderboard response: %w", err)
	}
	for _, u := range resp.Users {
		if strings.EqualFold(u.Pseudo, nickname) && u.AgentID != 0 {
			return u.toAgentInfo(), nil
		}
	}
	if len(resp.Users) > 0 && resp.Users[0].AgentID != 0 {
		return resp.Users[0].toAgentInfo(), nil
	}
	return AgentInfo{}, fmt.Errorf("no agentId found for nickname %q in puzzle %q", nickname, apiSlug)
}

type leaderboardUser struct {
	Pseudo  string  `json:"pseudo"`
	AgentID int64   `json:"agentId"`
	Rank    int     `json:"rank"`
	Score   float64 `json:"score"`
	League  struct {
		DivisionIndex int `json:"divisionIndex"`
	} `json:"league"`
}

func (u leaderboardUser) toAgentInfo() AgentInfo {
	return AgentInfo{
		AgentID:  u.AgentID,
		Rank:     u.Rank,
		Division: u.League.DivisionIndex,
		Score:    u.Score,
	}
}

// FindLastBattles returns the gameIds for the most recent battles played by
// the given agentId.
func (c *Client) FindLastBattles(agentID int64) ([]int64, error) {
	body, err := c.post(LastBattlesAPI, fmt.Sprintf("[%d,null]", agentID))
	if err != nil {
		return nil, fmt.Errorf("fetch last battles: %w", err)
	}
	var resp []struct {
		GameID int64 `json:"gameId"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse last battles: %w", err)
	}
	ids := make([]int64, 0, len(resp))
	for _, b := range resp {
		if b.GameID != 0 {
			ids = append(ids, b.GameID)
		}
	}
	return ids, nil
}

func (c *Client) post(endpoint, body string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, data)
	}
	return data, nil
}
