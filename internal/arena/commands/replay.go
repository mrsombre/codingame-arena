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

// replayFetcher is the minimal CodinGame client surface the download loop
// needs. Defined here (not in the codingame package) to keep the dependency
// arrow inward and let tests substitute a fake without spinning up HTTP.
type replayFetcher interface {
	FetchReplay(gameID int64) ([]byte, error)
}

type downloadOutcome int

const (
	downloadOutcomeSaved downloadOutcome = iota
	downloadOutcomeSkippedExisting
	downloadOutcomeFailed
)

type downloadResult struct {
	ID      int64
	Outcome downloadOutcome
	Detail  string
}

type downloadSummary struct {
	Total   int
	Saved   int
	Skipped int
	Failed  int
}

// replayBatchConfig bundles the knobs the download loop needs. Built from
// either ReplayGetOptions or ReplayLeaderboardOptions plus the per-command
// annotations layered into every saved replay.
type replayBatchConfig struct {
	IDs         []int64
	Annotations arena.ReplayAnnotations
	OutDir      string
	Limit       int
	Delay       time.Duration
	Force       bool
}

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
func ReplayGet(args []string, stdout io.Writer, factory arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	opts, err := parseReplayGetOptions(args, fs, v)
	if err != nil {
		return err
	}

	cfg := replayBatchConfig{
		IDs:    opts.IDs,
		OutDir: opts.OutDir,
		Limit:  opts.Limit,
		Delay:  opts.Delay,
		Force:  opts.Force,
		Annotations: arena.ReplayAnnotations{
			Blue:        opts.Username,
			League:      opts.League,
			Source:      arena.ReplaySourceGet,
			PuzzleID:    factory.PuzzleID(),
			PuzzleTitle: factory.PuzzleTitle(),
		},
	}
	return downloadReplays(codingame.New(), cfg, stdout)
}

// ReplayLeaderboard is the entry point for the "replay leaderboard"
// subcommand. It resolves a CodinGame leaderboard URL plus nickname into the
// player's last battles and downloads each replay as pretty-printed JSON.
func ReplayLeaderboard(args []string, stdout io.Writer, factory arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
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

	info, err := resolveAgent(client, store, apiSlug, opts.Username)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "player: %s -> agentId %d (rank %d, division %d)\n",
		opts.Username, info.AgentID, info.Rank, info.Division)

	gameIDs, err := client.FindLastBattles(info.AgentID)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "battles: %d\n", len(gameIDs))

	cfg := replayBatchConfig{
		IDs:    gameIDs,
		OutDir: opts.OutDir,
		Limit:  opts.Limit,
		Delay:  opts.Delay,
		Force:  opts.Force,
		Annotations: arena.ReplayAnnotations{
			Blue:   opts.Username,
			League: opts.League,
			Source: arena.ReplaySourceLeaderboard,
			Leaderboard: &arena.ReplayLeaderboardInfo{
				Rank:     info.Rank,
				Division: info.Division,
				Score:    info.Score,
			},
			PuzzleID:    factory.PuzzleID(),
			PuzzleTitle: factory.PuzzleTitle(),
		},
	}
	return downloadReplays(client, cfg, stdout)
}

// downloadReplays is the shared per-ID download orchestrator. Skipped-existing
// replays don't reset the inter-fetch delay; fetch errors are soft (counted
// as failed); prepare/write errors abort the batch.
func downloadReplays(fetcher replayFetcher, cfg replayBatchConfig, stdout io.Writer) error {
	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", cfg.OutDir, err)
	}
	results, err := runReplayDownloads(fetcher, cfg, stdout, time.Now)
	if err != nil {
		return err
	}
	return writeDownloadSummary(stdout, cfg.OutDir, results)
}

func runReplayDownloads(fetcher replayFetcher, cfg replayBatchConfig, stdout io.Writer, now func() time.Time) ([]downloadResult, error) {
	ids := cfg.IDs
	if cfg.Limit > 0 && len(ids) > cfg.Limit {
		ids = ids[:cfg.Limit]
	}

	results := make([]downloadResult, 0, len(ids))
	fetched := 0
	for i, id := range ids {
		if !cfg.Force && replayFileExists(cfg.OutDir, id) {
			result := downloadResult{ID: id, Outcome: downloadOutcomeSkippedExisting, Detail: "exists"}
			writeDownloadProgress(stdout, i+1, len(ids), result)
			results = append(results, result)
			continue
		}

		if fetched > 0 && cfg.Delay > 0 {
			time.Sleep(cfg.Delay)
		}

		result, err := downloadReplay(fetcher, cfg, id, now())
		if err != nil {
			return nil, err
		}
		fetched++
		writeDownloadProgress(stdout, i+1, len(ids), result)
		results = append(results, result)
	}
	return results, nil
}

// downloadReplay handles one target. Fetch errors map to a soft failed result
// so the batch keeps going; prepare/write errors return an error to abort.
func downloadReplay(fetcher replayFetcher, cfg replayBatchConfig, id int64, fetchedAt time.Time) (downloadResult, error) {
	body, err := fetcher.FetchReplay(id)
	if err != nil {
		return downloadResult{ID: id, Outcome: downloadOutcomeFailed, Detail: err.Error()}, nil
	}

	ann := cfg.Annotations
	ann.FetchedAt = fetchedAt
	body, err = arena.PrepareReplay(body, ann)
	if err != nil {
		return downloadResult{}, fmt.Errorf("prepare replay %d: %w", id, err)
	}

	path := replayFilePath(cfg.OutDir, id)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return downloadResult{}, fmt.Errorf("write %s: %w", path, err)
	}
	return downloadResult{ID: id, Outcome: downloadOutcomeSaved, Detail: fmt.Sprintf("%d bytes", len(body))}, nil
}

func replayFilePath(outDir string, id int64) string {
	return filepath.Join(outDir, fmt.Sprintf("%d.json", id))
}

func replayFileExists(outDir string, id int64) bool {
	_, err := os.Stat(replayFilePath(outDir, id))
	return err == nil
}

func writeDownloadProgress(stdout io.Writer, current, total int, r downloadResult) {
	switch r.Outcome {
	case downloadOutcomeSaved:
		_, _ = fmt.Fprintf(stdout, "[%d/%d] save %d (%s)\n", current, total, r.ID, r.Detail)
	case downloadOutcomeSkippedExisting:
		_, _ = fmt.Fprintf(stdout, "[%d/%d] skip %d (%s)\n", current, total, r.ID, r.Detail)
	case downloadOutcomeFailed:
		_, _ = fmt.Fprintf(stdout, "[%d/%d] fail %d: %s\n", current, total, r.ID, r.Detail)
	}
}

func writeDownloadSummary(stdout io.Writer, outDir string, results []downloadResult) error {
	s := summarizeDownloadResults(results)
	_, err := fmt.Fprintf(stdout, "done: %d saved, %d skipped, %d failed (out=%s)\n",
		s.Saved, s.Skipped, s.Failed, outDir)
	return err
}

func summarizeDownloadResults(results []downloadResult) downloadSummary {
	s := downloadSummary{Total: len(results)}
	for _, r := range results {
		switch r.Outcome {
		case downloadOutcomeSaved:
			s.Saved++
		case downloadOutcomeSkippedExisting:
			s.Skipped++
		case downloadOutcomeFailed:
			s.Failed++
		}
	}
	return s
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

// resolveAgent fetches the player's current leaderboard standing and refreshes
// the local cache. Always hits the API: rank/division change continuously, so
// stale cache entries would silently mislabel saved replays.
func resolveAgent(client *codingame.Client, store *db.DB, apiSlug, nickname string) (codingame.AgentInfo, error) {
	info, err := client.FindAgent(apiSlug, nickname)
	if err != nil {
		return codingame.AgentInfo{}, err
	}
	if err := store.Players.Save(apiSlug, nickname, info.AgentID); err != nil {
		return codingame.AgentInfo{}, fmt.Errorf("cache player: %w", err)
	}
	return info, nil
}
