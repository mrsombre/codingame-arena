package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
	"github.com/mrsombre/codingame-arena/internal/arena/server"
	"github.com/mrsombre/codingame-arena/viewer"
)

// Front is the entry point for the "front" subcommand. It serves the embedded
// viewer bundle over HTTP and listens on stdin for single-letter commands.
func Front(args []string, stdout io.Writer, factory arena.GameFactory, flags *pflag.FlagSet, v *viper.Viper) error {
	knownArgs, _, err := arena.SplitArgs(args, flags)
	if err != nil {
		return err
	}
	if err := flags.Parse(knownArgs); err != nil {
		return err
	}

	if v.GetBool("help") {
		extra := "API: GET /api/game, GET /api/games, GET /api/bots, GET /api/matches, GET /api/matches/{id},\n" +
			"     GET /api/replays, GET /api/replays/{id}, POST /api/run\n" +
			"Stdin keys: o<enter> open in default browser   q<enter> quit"
		_, err := fmt.Fprintln(stdout, arena.CommandUsage("serve", "Serve the embedded web viewer.", flags, extra))
		return err
	}

	port := v.GetInt("port")
	host := v.GetString("host")
	traceDir := v.GetString("trace-dir")
	replayDir := v.GetString("replay-dir")
	binDir := v.GetString("bin-dir")
	if port < 1 || port > 65535 {
		return fmt.Errorf("--port must be in 1..65535")
	}

	bots := scanBots(binDir)

	assets, err := fs.Sub(viewer.Assets, "dist")
	if err != nil {
		return fmt.Errorf("sub dist: %w", err)
	}
	handler := server.New(server.Options{
		Factory:   factory,
		Assets:    assets,
		TraceDir:  traceDir,
		ReplayDir: replayDir,
		Bots:      bots,
	})

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	url := fmt.Sprintf("http://%s", addr)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	watching, _ := resolveExe()
	printFrontUsage(stdout, url, traceDir, watching)

	binaryChanged := watchBinary(ctx, stdout)
	stdinCmds := readStdinCommands(ctx)

	for {
		select {
		case <-ctx.Done():
			_, _ = fmt.Fprintln(stdout, "shutting down...")
			return shutdown(httpServer)

		case err := <-serverErr:
			if err != nil {
				return fmt.Errorf("server: %w", err)
			}
			return nil

		case <-binaryChanged:
			_, _ = fmt.Fprintln(stdout, "\nbinary changed — restarting...")
			_ = shutdown(httpServer)
			return reexec()

		case cmd, ok := <-stdinCmds:
			if !ok {
				<-ctx.Done()
				return shutdown(httpServer)
			}
			switch cmd {
			case "o":
				if err := openBrowser(url); err != nil {
					_, _ = fmt.Fprintf(stdout, "open browser: %v\n", err)
				} else {
					_, _ = fmt.Fprintf(stdout, "opened %s\n", url)
				}
				printFrontUsage(stdout, url, traceDir, watching)
			case "q":
				_, _ = fmt.Fprintln(stdout, "shutting down...")
				return shutdown(httpServer)
			default:
				printFrontUsage(stdout, url, traceDir, watching)
			}
		}
	}
}

func readStdinCommands(ctx context.Context) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			select {
			case <-ctx.Done():
				return
			case out <- line:
			}
		}
	}()
	return out
}

func shutdown(server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

func printFrontUsage(w io.Writer, url, traceDir, watching string) {
	traceInfo := "(none)"
	if traceDir != "" {
		traceInfo = traceDir
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "  arena front  ready")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "  ➜  Local:     %s\n", url)
	_, _ = fmt.Fprintf(w, "  ➜  Trace:     %s\n", traceInfo)
	if watching != "" {
		_, _ = fmt.Fprintf(w, "  ➜  Watching:  %s\n", watching)
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "  Shortcuts")
	_, _ = fmt.Fprintln(w, "  press o + enter to open in browser")
	_, _ = fmt.Fprintln(w, "  press q + enter to quit")
	_, _ = fmt.Fprintln(w)
}

func resolveExe() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

// watchBinary watches the current executable for changes and signals on the
// returned channel. It watches the parent directory (not the file itself)
// because `go build` replaces the file atomically (delete + rename), which
// would remove fsnotify's inode watch.
func watchBinary(ctx context.Context, log io.Writer) <-chan struct{} {
	ch := make(chan struct{}, 1)

	exe, err := resolveExe()
	if err != nil {
		_, _ = fmt.Fprintf(log, "watch: cannot resolve executable: %v\n", err)
		return ch
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		_, _ = fmt.Fprintf(log, "watch: cannot create watcher: %v\n", err)
		return ch
	}

	dir := filepath.Dir(exe)
	base := filepath.Base(exe)
	if err := watcher.Add(dir); err != nil {
		_, _ = fmt.Fprintf(log, "watch: cannot watch %s: %v\n", dir, err)
		_ = watcher.Close()
		return ch
	}

	go func() {
		defer func() { _ = watcher.Close() }()
		// Debounce: builds may fire multiple events in quick succession.
		var debounce *time.Timer
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}
				if filepath.Base(ev.Name) != base {
					continue
				}
				if ev.Op&(fsnotify.Create|fsnotify.Write) == 0 {
					continue
				}
				if debounce != nil {
					debounce.Stop()
				}
				debounce = time.AfterFunc(300*time.Millisecond, func() {
					select {
					case ch <- struct{}{}:
					default:
					}
				})
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	return ch
}

// reexec replaces the current process with a fresh instance of the same binary.
func reexec() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("reexec: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("reexec: %w", err)
	}
	return syscall.Exec(exe, os.Args, os.Environ())
}

// scanBots reads binDir and returns absolute paths of executable files
// whose name contains "bot".
func scanBots(binDir string) []string {
	if binDir == "" {
		return nil
	}
	entries, err := os.ReadDir(binDir)
	if err != nil {
		return nil
	}
	var bots []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.Contains(strings.ToLower(e.Name()), "bot") {
			continue
		}
		abs, err := filepath.Abs(filepath.Join(binDir, e.Name()))
		if err != nil {
			continue
		}
		bots = append(bots, abs)
	}
	return bots
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}
