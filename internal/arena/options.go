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
	Timing      bool
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
	fs.String("game", "", "Active game (auto-detected when only one is registered)")
	return fs
}

// AddRunFlags registers flags used by the "run" subcommand on fs.
func AddRunFlags(fs *pflag.FlagSet) {
	fs.Int("simulations", 1, "Number of matches to run")
	fs.Int("parallel", runtime.NumCPU(), "Worker threads")
	fs.String("seed", "", "Base RNG seed (default: current time)")
	fs.String("seedx", "", "Seed increment per match (seed_i = seed + i*N)")
	fs.Bool("output-matches", false, "Include per-match results in JSON output")
	fs.Bool("debug", false, "Force one match, fixed sides, print debug to stderr")
	fs.Bool("timing", false, "Print per-turn timing to stderr")
	fs.Bool("no-swap", false, "Disable automatic side swapping")
	fs.String("trace-dir", "", "Write per-match JSON trace files")
	fs.Int("max-turns", 200, "Maximum turns per match")
	fs.String("p0-bin", "", "Player 0 binary (required)")
	fs.String("p1-bin", filepath.Clean("./bin/opponent"), "Player 1 binary")
	fs.Bool("verbose", false, "Output full JSON (default: short summary line)")
}

// AddSerializeFlags registers flags used by the "serialize" subcommand on fs.
func AddSerializeFlags(fs *pflag.FlagSet) {
	fs.String("seed", "", "RNG seed (required)")
	fs.Int("player", -1, "Player index (0 or 1)")
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

	parsed := ParsedArgs{
		BatchOptions: BatchOptions{
			Simulations:   v.GetInt("simulations"),
			Parallel:      v.GetInt("parallel"),
			OutputMatches: v.GetBool("output-matches"),
		},
		P0Bin:       v.GetString("p0-bin"),
		P1Bin:       v.GetString("p1-bin"),
		MaxTurns:    v.GetInt("max-turns"),
		TraceDir:    v.GetString("trace-dir"),
		Debug:       v.GetBool("debug"),
		Timing:      v.GetBool("timing"),
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

	if raw := v.GetString("seedx"); raw != "" {
		n, err := ParseSeed(raw)
		if err != nil {
			return ParsedArgs{}, fmt.Errorf("invalid integer for --seedx: %s", raw)
		}
		parsed.SeedIncrement = &n
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
	if parsed.SeedIncrement != nil && *parsed.SeedIncrement <= 0 {
		return ParsedArgs{}, fmt.Errorf("--seedx must be >= 1")
	}
	if !parsed.Help && parsed.P0Bin == "" {
		return ParsedArgs{}, fmt.Errorf("--p0-bin is required")
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

// Usage returns the top-level help text, including available games.
func Usage(games []string) string {
	return strings.TrimSpace(fmt.Sprintf(`Available games: %s

Usage: arena <command> [OPTIONS]

arena run - Run one or more match simulations against a player binary.
  --simulations <N>    Number of matches to run (default: 1)
  --parallel <N>       Number of worker threads (default: logical CPUs)
  --seed <N>           Base RNG seed (default: current time)
  --seedx <N>          Seed increment per match (seed_i = seed + i*N)
  --output-matches     Include per-match results in JSON output
  --trace-dir <PATH>   Write per-match JSON trace files
  --no-swap            Disable automatic side swapping
  --debug              Force one match, fixed sides, print debug to stderr
  --max-turns <N>      Maximum turns per match (default: 200)
  --p0-bin <PATH>      Player 0 binary (required)
  --p1-bin <PATH>      Player 1 binary (default: ./bin/opponent)
  --timing             Print per-turn timing to stderr
  --verbose            Output full JSON (default: short summary line)

arena serialize - Print initial global + first-frame inputs for a given seed.
  --seed <N>           RNG seed (required)
  --player <0|1>       Player index (required)

arena replay <url|id> - Download raw replay JSON from codingame.com.
  -o, --out <PATH>     Output path. Default: ./replays/replay-<id>.json.
                       Trailing "/" or existing dir → replay-<id>.json inside.
                       Otherwise treated as a file path (created/overwritten).

arena serve - Serve the embedded web viewer.
  --port <N>           HTTP port (default: 5757)
  --host <HOST>        Bind host (default: localhost)
  --trace-dir <PATH>   Directory with match trace JSON files (powers /api/matches)
  --bin-dir <PATH>     Directory to scan for bot binaries (default: ./bin)
  API: GET /api/game, GET /api/games, GET /api/bots, GET /api/matches, GET /api/matches/{id}, POST /api/run
  Stdin keys: o<enter> open in default browser   q<enter> quit

Common options:
  --game <NAME>        Active game (auto-detected when only one is registered)
  -h, --help           Show this help

Env vars: ARENA_<FLAG> (hyphens become underscores, e.g. ARENA_GAME, ARENA_SEED).
Config: arena.yml in current directory (e.g. game: winter-2026).

Unknown --key value flags are passed as game options to the engine factory.`, strings.Join(games, ", ")))
}

func ParseSeed(value string) (int64, error) {
	raw := strings.TrimPrefix(value, "seed=")
	return strconv.ParseInt(raw, 10, 64)
}
