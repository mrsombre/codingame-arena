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

// ReplayUsage returns the help text shown for `arena replay` without a
// recognized sub-subcommand.
func ReplayUsage() string {
	return `arena replay - Download raw replay JSON from codingame.com.

Usage: arena replay <subcommand> [OPTIONS]

Subcommands:
  get          Download one or more replays by ID/URL
  leaderboard  Download every replay from a player's last battles list

Use "arena help replay <subcommand>" for more information about a subcommand.

Env vars: ARENA_<FLAG> (hyphens become underscores, e.g. ARENA_GAME, ARENA_SEED).
Config: arena.yml in current directory (e.g. game: winter2026).`
}

// ReplayGetUsage returns the help text shown for `arena help replay get`.
func ReplayGetUsage(fs *pflag.FlagSet) string {
	return arena.CommandUsage(
		"replay get <username> <id|url>[,<id|url>...]",
		"Download raw replay JSON for one or more CodinGame games. <username> is "+
			"the player we are playing for; it is recorded as the top-level "+
			"\"blue\" field in every saved replay.",
		fs,
		"",
	)
}

// ReplayLeaderboardUsage returns the help text shown for
// `arena help replay leaderboard`.
func ReplayLeaderboardUsage(fs *pflag.FlagSet) string {
	return arena.CommandUsage(
		"replay leaderboard <username> <puzzle-url|slug>",
		"Download every replay from a player's last battles list. <username> is "+
			"the player we are playing for; it is recorded as the top-level "+
			"\"blue\" field in every saved replay.",
		fs,
		"",
	)
}

// ReplayGet is the entry point for the "replay get" subcommand. It downloads
// the raw replay JSON for one or more CodinGame games, strips the unused
// top-level viewer payload, and writes each replay back as pretty-printed JSON.
func ReplayGet(args []string, stdout io.Writer, _ arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	opts, err := parseReplayGetOptions(args, fs, v)
	if err != nil {
		return err
	}

	client := codingame.New()
	return downloadReplays(client, opts.IDs, opts.Username, opts.League, opts.OutDir, opts.Limit, opts.Delay, opts.Force, stdout)
}

// ReplayLeaderboard is the entry point for the "replay leaderboard"
// subcommand. It resolves a CodinGame leaderboard URL plus nickname into the
// player's last battles and downloads each replay as pretty-printed JSON.
func ReplayLeaderboard(args []string, stdout io.Writer, _ arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	opts, err := parseReplayLeaderboardOptions(args, fs, v)
	if err != nil {
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

	agentID, playerHit, err := resolveAgent(client, store, apiSlug, opts.Username)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "player: %s -> agentId %d%s\n", opts.Username, agentID, cacheTag(playerHit))

	gameIDs, err := client.FindLastBattles(agentID)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "battles: %d\n", len(gameIDs))

	return downloadReplays(client, gameIDs, opts.Username, opts.League, opts.OutDir, opts.Limit, opts.Delay, opts.Force, stdout)
}

// downloadReplays runs the shared per-ID download loop: skip-if-exists (unless
// force is set), inter-request delay, soft failure, and a final summary line.
// blue is the username we are playing for; when non-empty it is written as
// the top-level "blue" field of every saved replay JSON. league, when
// non-zero, is written as the top-level "league" field.
func downloadReplays(client *codingame.Client, ids []int64, blue string, league int, outDir string, limit int, delay time.Duration, force bool, stdout io.Writer) error {
	if limit > 0 && len(ids) > limit {
		ids = ids[:limit]
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("create %s: %w", outDir, err)
	}

	var saved, skipped, failed int
	first := true
	for i, id := range ids {
		path := filepath.Join(outDir, fmt.Sprintf("%d.json", id))
		if !force {
			if _, err := os.Stat(path); err == nil {
				skipped++
				_, _ = fmt.Fprintf(stdout, "[%d/%d] skip %d (exists)\n", i+1, len(ids), id)
				continue
			}
		}

		if !first && delay > 0 {
			time.Sleep(delay)
		}
		first = false

		body, err := client.FetchReplay(id)
		if err != nil {
			failed++
			_, _ = fmt.Fprintf(stdout, "[%d/%d] fail %d: %v\n", i+1, len(ids), id, err)
			continue
		}
		body, err = arena.PrepareReplay(body, blue, league)
		if err != nil {
			return fmt.Errorf("prepare replay %d: %w", id, err)
		}
		if err := os.WriteFile(path, body, 0644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		saved++
		_, _ = fmt.Fprintf(stdout, "[%d/%d] save %d (%d bytes)\n", i+1, len(ids), id, len(body))
	}

	_, _ = fmt.Fprintf(stdout, "done: %d saved, %d skipped, %d failed (out=%s)\n", saved, skipped, failed, outDir)
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
