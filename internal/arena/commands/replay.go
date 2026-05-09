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
	downloadOutcomeSkippedPuzzle
	downloadOutcomeFailed
)

type downloadResult struct {
	ID      int64
	Outcome downloadOutcome
	Detail  string
}

type downloadSummary struct {
	Total           int
	Saved           int
	SkippedExisting int
	SkippedPuzzle   int
	Failed          int
}

// replayBatchConfig bundles the knobs the download + auto-convert loop needs.
// Built from ReplayOptions plus the per-mode annotations layered into every
// saved replay. Factory is required for auto-convert; tests pass nil to skip
// the conversion step.
type replayBatchConfig struct {
	IDs         []int64
	Annotations arena.ReplayAnnotations
	OutDir      string
	TraceDir    string
	Factory     arena.GameFactory
	Limit       int
	Delay       time.Duration
	Force       bool
}

// ReplayUsage returns the help text shown for `arena help replay`.
func ReplayUsage(fs *pflag.FlagSet) string {
	extra := `Positional args:
  arena replay <game> <username> [<id|url>[,<id|url>...]]
    <game>      engine slug (e.g. winter2026, spring2020); selects which
                CodinGame leaderboard slug + puzzleId to use.
    <username>  CodinGame nickname we are playing for. Stamped into every
                saved replay as the top-level "blue" field so analyze and
                the viewer know which side is "us".
    <id|url>    optional: zero or more replay ids (numeric) or full replay
                URLs ending in an id. Pass them as separate args, comma-
                separated within one arg, or both.

Modes:
  No ids → leaderboard mode. Resolves the active game's leaderboard slug,
           looks up <username>'s agentId on it, and downloads every replay
           from that player's last-battles list. Slug + agentId lookups are
           cached in db.sqlite3.
  Ids    → get mode. Fetches only the listed games from codingame.com.

Auto-convert:
  Each freshly-saved replay is immediately converted to a verified arena
  trace under the same id: replays/<id>.json → traces/replay-<id>.json.
  This is NOT a separate convert command. The conversion re-runs the engine
  with the recorded player moves and verifies final scores, ranks, and turn
  counts against the replay. Verifier disagreements still write a trace
  (tagged MISMATCH so analyze can compare engine vs replay); pre-run
  failures (missing seed, unknown blue) are skipped with no trace.

Skipping rules:
  Replays already on disk are skipped (download + convert). Pass -f/--force
  to refetch AND reconvert, overwriting both files. To regenerate ONLY a
  missing trace without redownloading: delete the trace file and re-run
  without -f — the existing replay JSON is kept and only the trace is
  rebuilt.

Output (per replay):
  [i/N] save  <id> (<bytes>)             freshly downloaded
  [i/N] trace <id> (league=L turns=T scores=A:B)   trace written
  [i/N] trace <id> MISMATCH (...)        verifier disagreement, trace still written
  [i/N] skip  <id> (exists | puzzleId M != N)
  [i/N] fail  <id>: <error>
  done:   <saved>/<skipped-existing>/<skipped-puzzle>/<failed>  (out=...)
  traces: <saved>/<saved-mismatch>/<skipped-existing>/<skipped-mismatch>/<failed>  (out=...)

Files:
  --out      → replays/<gameId>.json (raw replay payload + arena annotations)
  --trace-dir → traces/replay-<gameId>.json (the verified trace)
  Both feed into ` + "`arena analyze <game>`" + ` and the web viewer (` + "`arena serve`" + `).`
	return arena.CommandUsage(
		"replay <game> <username> [<id|url>...]",
		"Download CodinGame replays for a player and auto-convert each to a verified arena trace.",
		fs,
		extra,
	)
}

// Replay is the entry point for the "replay" command. With no IDs it
// downloads every replay from the player's last battles list on the active
// game's leaderboard; with one or more IDs/URLs it downloads only those games.
// Each freshly-downloaded replay is immediately converted to a trace file.
func Replay(args []string, stdout io.Writer, factory arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	opts, err := parseReplayOptions(args, fs, v)
	if err != nil {
		return err
	}

	if len(opts.IDs) > 0 {
		return replayByIDs(opts, factory, stdout)
	}
	return replayFromLeaderboard(opts, factory, stdout)
}

func replayByIDs(opts ReplayOptions, factory arena.GameFactory, stdout io.Writer) error {
	cfg := replayBatchConfig{
		IDs:      opts.IDs,
		OutDir:   opts.OutDir,
		TraceDir: opts.TraceDir,
		Factory:  factory,
		Limit:    opts.Limit,
		Delay:    opts.Delay,
		Force:    opts.Force,
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

func replayFromLeaderboard(opts ReplayOptions, factory arena.GameFactory, stdout io.Writer) error {
	slug := factory.LeaderboardSlug()
	if slug == "" {
		return fmt.Errorf("game %q has no leaderboard slug configured", factory.Name())
	}

	store, err := db.Open("")
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	client := codingame.New()

	apiSlug, puzzleHit, err := resolvePuzzle(client, store, slug)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "puzzle: %s -> %s%s\n", slug, apiSlug, cacheTag(puzzleHit))

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
		IDs:      gameIDs,
		OutDir:   opts.OutDir,
		TraceDir: opts.TraceDir,
		Factory:  factory,
		Limit:    opts.Limit,
		Delay:    opts.Delay,
		Force:    opts.Force,
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
// as failed); prepare/write errors abort the batch. Each freshly-saved replay
// is fed through convertReplay before the loop moves on (skipped when
// cfg.Factory is nil — only tests do that).
func downloadReplays(fetcher replayFetcher, cfg replayBatchConfig, stdout io.Writer) error {
	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", cfg.OutDir, err)
	}
	if cfg.Factory != nil {
		if err := os.MkdirAll(cfg.TraceDir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", cfg.TraceDir, err)
		}
	}
	dlResults, cvResults, err := runReplayDownloads(fetcher, cfg, stdout, time.Now)
	if err != nil {
		return err
	}
	if err := writeDownloadSummary(stdout, cfg.OutDir, dlResults); err != nil {
		return err
	}
	if cfg.Factory != nil {
		if err := writeAutoConvertSummary(stdout, cfg.TraceDir, cvResults); err != nil {
			return err
		}
	}
	return nil
}

func runReplayDownloads(fetcher replayFetcher, cfg replayBatchConfig, stdout io.Writer, now func() time.Time) ([]downloadResult, []convertResult, error) {
	ids := cfg.IDs
	if cfg.Limit > 0 && len(ids) > cfg.Limit {
		ids = ids[:cfg.Limit]
	}

	dlResults := make([]downloadResult, 0, len(ids))
	cvResults := make([]convertResult, 0, len(ids))
	fetched := 0
	for i, id := range ids {
		var dl downloadResult
		if !cfg.Force && replayFileExists(cfg.OutDir, id) {
			dl = downloadResult{ID: id, Outcome: downloadOutcomeSkippedExisting, Detail: "exists"}
		} else {
			if fetched > 0 && cfg.Delay > 0 {
				time.Sleep(cfg.Delay)
			}
			var err error
			dl, err = downloadReplay(fetcher, cfg, id, now())
			if err != nil {
				return nil, nil, err
			}
			fetched++
		}
		writeDownloadProgress(stdout, i+1, len(ids), dl)
		dlResults = append(dlResults, dl)

		if cfg.Factory == nil ||
			dl.Outcome == downloadOutcomeFailed ||
			dl.Outcome == downloadOutcomeSkippedPuzzle {
			continue
		}
		// Convert whenever a replay file is on disk. Force-overwrite the
		// trace when the download itself was a fresh save (the just-written
		// replay supersedes any stale trace); otherwise honor the user's
		// --force flag, so `delete traces && rerun replay` regenerates only
		// the missing traces without re-downloading anything.
		convertForce := cfg.Force || dl.Outcome == downloadOutcomeSaved
		cv := autoConvertReplay(cfg.Factory, cfg.TraceDir, convertReplayTarget{
			ID:   id,
			Path: replayFilePath(cfg.OutDir, id),
		}, convertForce)
		writeAutoConvertProgress(stdout, i+1, len(ids), cv)
		cvResults = append(cvResults, cv)
	}
	return dlResults, cvResults, nil
}

// downloadReplay handles one target. Fetch errors map to a soft failed result
// so the batch keeps going; prepare/write errors return an error to abort.
// A non-zero source puzzleId that disagrees with the selected game maps to
// a soft "skipped-puzzle" result — nothing is written, since saving a
// cross-game replay would silently pollute the per-game replays/ tree.
func downloadReplay(fetcher replayFetcher, cfg replayBatchConfig, id int64, fetchedAt time.Time) (downloadResult, error) {
	body, err := fetcher.FetchReplay(id)
	if err != nil {
		return downloadResult{ID: id, Outcome: downloadOutcomeFailed, Detail: err.Error()}, nil
	}

	ann := cfg.Annotations
	if ann.PuzzleID != 0 {
		if sourcePID, ok := arena.PeekReplayPuzzleID(body); ok && sourcePID != 0 && sourcePID != ann.PuzzleID {
			return downloadResult{
				ID:      id,
				Outcome: downloadOutcomeSkippedPuzzle,
				Detail:  fmt.Sprintf("puzzleId %d != %d", sourcePID, ann.PuzzleID),
			}, nil
		}
	}

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

// autoConvertReplay runs convertReplay and folds I/O errors into a soft Failed
// result so a single bad conversion does not abort the batch. force is true
// when a fresh replay was just saved (the trace must follow the new replay) or
// the user passed --force; otherwise convertReplay's existing-trace check
// keeps already-converted files untouched.
func autoConvertReplay(factory arena.GameFactory, traceDir string, target convertReplayTarget, force bool) convertResult {
	res, err := convertReplay(factory, ConvertOptions{TraceDir: traceDir, Force: force}, target)
	if err != nil {
		return convertResult{
			Target:  target,
			Outcome: convertOutcomeFailed,
			Detail:  err.Error(),
		}
	}
	return res
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
	case downloadOutcomeSkippedExisting, downloadOutcomeSkippedPuzzle:
		_, _ = fmt.Fprintf(stdout, "[%d/%d] skip %d (%s)\n", current, total, r.ID, r.Detail)
	case downloadOutcomeFailed:
		_, _ = fmt.Fprintf(stdout, "[%d/%d] fail %d: %s\n", current, total, r.ID, r.Detail)
	}
}

func writeDownloadSummary(stdout io.Writer, outDir string, results []downloadResult) error {
	s := summarizeDownloadResults(results)
	_, err := fmt.Fprintf(stdout, "done: %d saved, %d skipped-existing, %d skipped-puzzle, %d failed (out=%s)\n",
		s.Saved, s.SkippedExisting, s.SkippedPuzzle, s.Failed, outDir)
	return err
}

func summarizeDownloadResults(results []downloadResult) downloadSummary {
	s := downloadSummary{Total: len(results)}
	for _, r := range results {
		switch r.Outcome {
		case downloadOutcomeSaved:
			s.Saved++
		case downloadOutcomeSkippedExisting:
			s.SkippedExisting++
		case downloadOutcomeSkippedPuzzle:
			s.SkippedPuzzle++
		case downloadOutcomeFailed:
			s.Failed++
		}
	}
	return s
}

// writeAutoConvertProgress prints one status line per converted replay. The
// "trace" verb (rather than convert.go's "save") is used so the line is
// visually distinct from the preceding download's "save" line for the same ID.
func writeAutoConvertProgress(stdout io.Writer, current, total int, r convertResult) {
	switch r.Outcome {
	case convertOutcomeSaved:
		_, _ = fmt.Fprintf(stdout, "[%d/%d] trace %d (%s)\n", current, total, r.Target.ID, r.Detail)
	case convertOutcomeSavedMismatch:
		_, _ = fmt.Fprintf(stdout, "[%d/%d] trace %d MISMATCH (%s)\n", current, total, r.Target.ID, r.Detail)
	case convertOutcomeSkippedExisting, convertOutcomeSkippedPuzzle, convertOutcomeSkippedMismatch:
		_, _ = fmt.Fprintf(stdout, "[%d/%d] skip %d (%s)\n", current, total, r.Target.ID, r.Detail)
	case convertOutcomeFailed:
		_, _ = fmt.Fprintf(stdout, "[%d/%d] fail %d: %s\n", current, total, r.Target.ID, r.Detail)
	}
}

// writeAutoConvertSummary omits the convertOutcomeSkippedPuzzle column from
// the trace summary line — wrong-puzzle replays are now rejected at download
// time (the download summary's skipped-puzzle counter), so the auto-convert
// path here only ever sees files that already passed the puzzleId check.
func writeAutoConvertSummary(stdout io.Writer, traceDir string, results []convertResult) error {
	s := summarizeConvertResults(results)
	_, err := fmt.Fprintf(stdout, "traces: %d saved, %d saved-mismatch, %d skipped-existing, %d skipped-mismatch, %d failed (out=%s)\n",
		s.Saved, s.SavedMismatch, s.SkippedExisting, s.SkippedMismatch, s.Failed, traceDir)
	return err
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
