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
	"runtime"
	"strings"
	"syscall"
	"time"

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
		_, err := fmt.Fprintln(stdout, arena.Usage(factory))
		return err
	}

	port := v.GetInt("port")
	host := v.GetString("host")
	traceDir := v.GetString("trace-dir")
	if port < 1 || port > 65535 {
		return fmt.Errorf("--port must be in 1..65535")
	}

	assets, err := fs.Sub(viewer.Assets, "dist")
	if err != nil {
		return fmt.Errorf("sub dist: %w", err)
	}
	handler := server.New(server.Options{
		Factory:  factory,
		Assets:   assets,
		TraceDir: traceDir,
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

	printFrontUsage(stdout, url, traceDir)

	stdinCmds := readStdinCommands(ctx)

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(stdout, "shutting down...")
			return shutdown(httpServer)

		case err := <-serverErr:
			if err != nil {
				return fmt.Errorf("server: %w", err)
			}
			return nil

		case cmd, ok := <-stdinCmds:
			if !ok {
				<-ctx.Done()
				return shutdown(httpServer)
			}
			switch cmd {
			case "o":
				if err := openBrowser(url); err != nil {
					fmt.Fprintf(stdout, "open browser: %v\n", err)
				} else {
					fmt.Fprintf(stdout, "opened %s\n", url)
				}
				printFrontUsage(stdout, url, traceDir)
			case "q":
				fmt.Fprintln(stdout, "shutting down...")
				return shutdown(httpServer)
			default:
				printFrontUsage(stdout, url, traceDir)
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

func printFrontUsage(w io.Writer, url, traceDir string) {
	traceInfo := "(none — /api/matches returns [])"
	if traceDir != "" {
		traceInfo = traceDir
	}
	fmt.Fprintf(w, `arena front — viewer server
  url:       %s
  trace-dir: %s
  keys:      o<enter> open in browser   q<enter> quit
`, url, traceInfo)
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
