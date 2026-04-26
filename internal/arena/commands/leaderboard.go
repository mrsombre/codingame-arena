package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
	"github.com/mrsombre/codingame-arena/internal/arena/codingame"
	"github.com/mrsombre/codingame-arena/internal/arena/db"
)

// Leaderboard is the entry point for the "leaderboard" subcommand. It resolves
// a CodinGame leaderboard URL plus nickname into the player's last battles
// and downloads each replay's raw JSON to disk.
func Leaderboard(args []string, stdout io.Writer, _ arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	opts, err := parseLeaderboardOptions(args, fs, v)
	if err != nil {
		return err
	}

	if opts.Help {
		_, err := fmt.Fprintln(stdout, arena.CommandUsage(
			"leaderboard <leaderboard-url> <nickname>",
			"Download every replay from a player's last battles list.",
			fs,
			"",
		))
		return err
	}

	store, err := db.Open("")
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	client := codingame.New()

	apiSlug, puzzleHit, err := resolvePuzzle(client, store, opts.Slug)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "puzzle: %s -> %s%s\n", opts.Slug, apiSlug, cacheTag(puzzleHit))

	agentID, playerHit, err := resolveAgent(client, store, apiSlug, opts.Nickname)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "player: %s -> agentId %d%s\n", opts.Nickname, agentID, cacheTag(playerHit))

	gameIDs, err := client.FindLastBattles(agentID)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "battles: %d\n", len(gameIDs))

	if opts.Limit > 0 && len(gameIDs) > opts.Limit {
		gameIDs = gameIDs[:opts.Limit]
	}

	if err := os.MkdirAll(opts.OutDir, 0755); err != nil {
		return fmt.Errorf("create %s: %w", opts.OutDir, err)
	}

	var saved, skipped, failed int
	first := true
	for i, id := range gameIDs {
		path := filepath.Join(opts.OutDir, fmt.Sprintf("%d.json", id))
		if _, err := os.Stat(path); err == nil {
			skipped++
			_, _ = fmt.Fprintf(stdout, "[%d/%d] skip %d (exists)\n", i+1, len(gameIDs), id)
			continue
		}

		if !first && opts.Delay > 0 {
			time.Sleep(opts.Delay)
		}
		first = false

		body, err := client.FetchReplay(id)
		if err != nil {
			failed++
			_, _ = fmt.Fprintf(stdout, "[%d/%d] fail %d: %v\n", i+1, len(gameIDs), id, err)
			continue
		}
		if err := os.WriteFile(path, body, 0644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		saved++
		_, _ = fmt.Fprintf(stdout, "[%d/%d] save %d (%d bytes)\n", i+1, len(gameIDs), id, len(body))
	}

	_, _ = fmt.Fprintf(stdout, "done: %d saved, %d skipped, %d failed (out=%s)\n", saved, skipped, failed, opts.OutDir)
	return nil
}

// cacheTag returns a short " (cached)" suffix when hit is true.
func cacheTag(hit bool) string {
	if hit {
		return " (cached)"
	}
	return ""
}

// resolvePuzzle reads from cache, falling back to a CodinGame API lookup and
// persisting the result.
func resolvePuzzle(client *codingame.Client, store *db.DB, prettyID string) (string, bool, error) {
	if cached, err := store.Puzzles.Find(prettyID); err != nil {
		return "", false, err
	} else if cached != nil {
		return cached.LeaderboardID, true, nil
	}
	apiSlug, err := client.ResolvePuzzle(prettyID)
	if err != nil {
		return "", false, err
	}
	if err := store.Puzzles.Save(prettyID, apiSlug); err != nil {
		return "", false, fmt.Errorf("cache puzzle: %w", err)
	}
	return apiSlug, false, nil
}

// resolveAgent reads from cache, falling back to a CodinGame leaderboard
// search and persisting the result.
func resolveAgent(client *codingame.Client, store *db.DB, apiSlug, nickname string) (int64, bool, error) {
	if cached, err := store.Players.Find(apiSlug, nickname); err != nil {
		return 0, false, err
	} else if cached != nil {
		return cached.AgentID, true, nil
	}
	agentID, err := client.FindAgent(apiSlug, nickname)
	if err != nil {
		return 0, false, err
	}
	if err := store.Players.Save(apiSlug, nickname, agentID); err != nil {
		return 0, false, fmt.Errorf("cache player: %w", err)
	}
	return agentID, false, nil
}
