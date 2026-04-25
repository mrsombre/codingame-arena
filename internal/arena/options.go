package arena

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ParsedArgs holds all parsed CLI arguments.
type ParsedArgs struct {
	BatchOptions
	P0Bin       string
	P1Bin       string
	MaxTurns    int
	TraceDir    string
	Debug       bool
	NoSwap      bool
	Verbose     bool
	Help        bool
	GameOptions map[string]string
}

// NewBaseFlagSet returns a flag set pre-populated with flags shared by
// every subcommand. Subcommands register their own flags on top via
// AddRunFlags / AddSerializeFlags.
func NewBaseFlagSet(name string) *pflag.FlagSet {
	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)
	fs.SortFlags = false
	fs.SetOutput(io.Discard)
	fs.BoolP("help", "h", false, "Show this help")
	fs.String("game", "", "Active game")
	return fs
}

// AddRunFlags registers flags used by the "run" subcommand on fs.
func AddRunFlags(fs *pflag.FlagSet) {
	fs.StringP("league", "l", "", "League level (default: game-specific)")
	fs.IntP("simulations", "n", 100, "Number of matches to run")
	fs.IntP("parallel", "p", runtime.NumCPU(), "Worker threads")
	fs.StringP("seed", "s", "", "Base RNG seed (default: current time)")
	fs.Int("seedx", 1, "Seed increment per match (seed_i = seed + i*N)")
	fs.Bool("output-matches", false, "Include per-match results in JSON output")
	fs.Bool("debug", false, "Force one match, fixed sides, print debug to stderr")
	fs.Bool("no-swap", false, "Disable automatic side swapping")
	fs.String("trace-dir", "", "Write per-match JSON trace files")
	fs.Int("max-turns", 200, "Maximum turns per match")
	fs.String("p0", "", "Player 0 binary (required)")
	fs.String("p1", filepath.Clean("./bin/opponent"), "Player 1 binary")
	fs.Bool("verbose", false, "Output full JSON (default: short summary line)")
}

// AddSerializeFlags registers flags used by the "serialize" subcommand on fs.
func AddSerializeFlags(fs *pflag.FlagSet) {
	fs.StringP("league", "l", "", "League level (default: game-specific)")
	fs.StringP("seed", "s", "", "RNG seed (required)")
	fs.Int("player", 0, "Player index (0 or 1)")
}

// AddReplayFlags registers flags used by the "replay" subcommand on fs.
func AddReplayFlags(fs *pflag.FlagSet) {
	fs.StringP("out", "o", "", "Output path. Empty → ./replays/replay-<id>.json. Trailing '/' or existing directory → replay-<id>.json inside. Otherwise → a file at that path.")
}

// AddFrontFlags registers flags used by the "front" subcommand on fs.
func AddFrontFlags(fs *pflag.FlagSet) {
	fs.Int("port", 5757, "HTTP port")
	fs.String("host", "localhost", "Bind host")
	fs.String("trace-dir", "./matches", "Directory with match trace JSON files (powers /api/matches)")
	fs.String("replay-dir", "./replays", "Directory with CodinGame replay JSON files (powers /api/replays)")
	fs.String("bin-dir", "./bin", "Directory to scan for bot binaries (powers /api/bots)")
}

func ParseRunArgs(args []string, fs *pflag.FlagSet, v *viper.Viper) (ParsedArgs, error) {
	knownArgs, gameOptions, err := SplitArgs(args, fs)
	if err != nil {
		return ParsedArgs{}, err
	}
	if err := fs.Parse(knownArgs); err != nil {
		return ParsedArgs{}, err
	}

	MergeConfigGameOptions(v, fs, gameOptions)
	injectLeague(v, gameOptions)

	parsed := ParsedArgs{
		BatchOptions: BatchOptions{
			Simulations:   v.GetInt("simulations"),
			Parallel:      v.GetInt("parallel"),
			SeedIncrement: int64(v.GetInt("seedx")),
			OutputMatches: v.GetBool("output-matches"),
		},
		P0Bin:       v.GetString("p0"),
		P1Bin:       v.GetString("p1"),
		MaxTurns:    v.GetInt("max-turns"),
		TraceDir:    v.GetString("trace-dir"),
		Debug:       v.GetBool("debug"),
		NoSwap:      v.GetBool("no-swap"),
		Verbose:     v.GetBool("verbose"),
		Help:        v.GetBool("help"),
		GameOptions: gameOptions,
	}

	if raw := v.GetString("seed"); raw != "" {
		n, err := ParseSeed(raw)
		if err != nil {
			return ParsedArgs{}, fmt.Errorf("invalid integer for --seed: %s", raw)
		}
		parsed.Seed = n
	} else {
		parsed.Seed = time.Now().UnixNano()
	}

	if parsed.Simulations < 1 {
		return ParsedArgs{}, fmt.Errorf("--simulations must be >= 1")
	}
	if parsed.Parallel < 1 {
		return ParsedArgs{}, fmt.Errorf("--parallel must be >= 1")
	}
	if parsed.MaxTurns < 1 {
		return ParsedArgs{}, fmt.Errorf("--max-turns must be >= 1")
	}
	if parsed.SeedIncrement < 1 {
		return ParsedArgs{}, fmt.Errorf("--seedx must be >= 1")
	}
	if !parsed.Help && parsed.P0Bin == "" {
		return ParsedArgs{}, fmt.Errorf("--p0 is required")
	}
	if parsed.Debug {
		parsed.Simulations = 1
		parsed.Parallel = 1
	}

	return parsed, nil
}

// NewViper returns a viper instance bound to the given flag set, configured
// with the ARENA_ env prefix, and populated from an "arena" config file
// (yaml/json/toml) in the current directory if present. A missing file is
// not an error; a malformed one is.
func NewViper(fs *pflag.FlagSet) (*viper.Viper, error) {
	v := viper.New()
	v.SetEnvPrefix("ARENA")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	v.SetConfigName("arena")
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, err
		}
	}

	if err := v.BindPFlags(fs); err != nil {
		return nil, err
	}
	return v, nil
}

// SplitArgs separates args known to fs from unknown --key value pairs,
// returning the unknown pairs as a game options map.
func SplitArgs(args []string, fs *pflag.FlagSet) ([]string, map[string]string, error) {
	known := make([]string, 0, len(args))
	gameOptions := make(map[string]string)

	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "--") || a == "--" {
			known = append(known, a)
			continue
		}

		trimmed := strings.TrimPrefix(a, "--")
		name := trimmed
		var value string
		hasInline := false
		if eq := strings.Index(trimmed, "="); eq >= 0 {
			name = trimmed[:eq]
			value = trimmed[eq+1:]
			hasInline = true
		}

		if f := fs.Lookup(name); f != nil {
			known = append(known, a)
			if !hasInline && !isBoolFlag(f) {
				if i+1 >= len(args) {
					return nil, nil, fmt.Errorf("missing value for --%s", name)
				}
				known = append(known, args[i+1])
				i++
			}
			continue
		}

		if !hasInline {
			if i+1 >= len(args) {
				return nil, nil, fmt.Errorf("missing value for --%s", name)
			}
			value = args[i+1]
			i++
		}
		gameOptions[name] = value
	}

	return known, gameOptions, nil
}

func isBoolFlag(f *pflag.Flag) bool {
	bf, ok := f.Value.(interface{ IsBoolFlag() bool })
	return ok && bf.IsBoolFlag()
}

// MergeConfigGameOptions copies viper keys that aren't bound to known flags
// into gameOptions, leaving existing (CLI-provided) entries untouched.
func MergeConfigGameOptions(v *viper.Viper, fs *pflag.FlagSet, gameOptions map[string]string) {
	for _, key := range v.AllKeys() {
		if fs.Lookup(key) != nil {
			continue
		}
		if _, ok := gameOptions[key]; ok {
			continue
		}
		raw := v.Get(key)
		if raw == nil {
			continue
		}
		gameOptions[key] = fmt.Sprintf("%v", raw)
	}
}

// InjectLeague copies the --league flag value into gameOptions if set.
// Exported for use by subcommands that build gameOptions outside ParseRunArgs.
func InjectLeague(v *viper.Viper, gameOptions map[string]string) {
	injectLeague(v, gameOptions)
}

func injectLeague(v *viper.Viper, gameOptions map[string]string) {
	if league := v.GetString("league"); league != "" {
		if _, exists := gameOptions["league"]; !exists {
			gameOptions["league"] = league
		}
	}
}

// Usage returns the top-level help text, listing available commands.
func Usage(games []string) string {
	return strings.TrimSpace(fmt.Sprintf(`Available games: %s

Usage: arena <command> [OPTIONS]

Commands:
  run          Run one or more match simulations against a player binary
  serialize    Print initial game input for first turn for a given seed
  replay       Download raw replay JSON from codingame.com
  serve        Serve the embedded web viewer

Use "arena <command> --help" for more information about a command.

Env vars: ARENA_<FLAG> (hyphens become underscores, e.g. ARENA_GAME, ARENA_SEED).
Config: arena.yml in current directory (e.g. game: winter2026).

Unknown --key value flags are passed as game options to the engine factory.`, strings.Join(games, ", ")))
}

// CommandUsage returns help text for a specific subcommand using fs.FlagUsages().
func CommandUsage(command, description string, fs *pflag.FlagSet, extra string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "arena %s - %s\n\nOptions:\n", command, description)
	sb.WriteString(fs.FlagUsages())
	if extra != "" {
		sb.WriteString("\n")
		sb.WriteString(extra)
		sb.WriteString("\n")
	}
	sb.WriteString("\nEnv vars: ARENA_<FLAG> (hyphens become underscores, e.g. ARENA_GAME, ARENA_SEED).\n")
	sb.WriteString("Config: arena.yml in current directory (e.g. game: winter2026).\n")
	sb.WriteString("Unknown --key value flags are passed as game options to the engine factory.")
	return sb.String()
}

func ParseSeed(value string) (int64, error) {
	raw := strings.TrimPrefix(value, "seed=")
	return strconv.ParseInt(raw, 10, 64)
}
