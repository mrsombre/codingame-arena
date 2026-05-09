package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunPrintsTopLevelHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Usage: arena <command> [<game>] [OPTIONS]") {
		t.Fatalf("stdout missing top-level usage:\n%s", stdout.String())
	}
}

func TestRunReportsUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"missing"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), `unknown command "missing"`) {
		t.Fatalf("stderr missing unknown-command error: %q", stderr.String())
	}
}

func TestNormalizeCommandSupportsTopLevelHelpFlags(t *testing.T) {
	for _, arg := range []string{"--help", "-h"} {
		command, rest := normalizeCommand([]string{arg})
		if command != "help" {
			t.Fatalf("command for %s = %q, want help", arg, command)
		}
		if len(rest) != 0 {
			t.Fatalf("rest for %s = %v, want empty", arg, rest)
		}
	}
}

func TestSelectCommandRoutesReplay(t *testing.T) {
	spec, path, rest, err := selectCommand("replay", []string{"winter2026", "mrsombre", "123"})
	if err != nil {
		t.Fatalf("selectCommand returned error: %v", err)
	}
	if spec.handler == nil {
		t.Fatal("handler is nil")
	}
	if !spec.needsFactory {
		t.Fatal("needsFactory = false, want true")
	}
	if path != "replay" {
		t.Fatalf("path = %q, want replay", path)
	}
	if got, want := strings.Join(rest, " "), "winter2026 mrsombre 123"; got != want {
		t.Fatalf("rest = %q, want %q", got, want)
	}
}

func TestSelectCommandDescendsIntoGameSerialize(t *testing.T) {
	spec, path, rest, err := selectCommand("game", []string{"serialize", "winter2026", "12345"})
	if err != nil {
		t.Fatalf("selectCommand returned error: %v", err)
	}
	if spec.handler == nil {
		t.Fatal("handler is nil")
	}
	if !spec.needsFactory {
		t.Fatal("needsFactory = false, want true")
	}
	if path != "game serialize" {
		t.Fatalf("path = %q, want \"game serialize\"", path)
	}
	if got, want := strings.Join(rest, " "), "winter2026 12345"; got != want {
		t.Fatalf("rest = %q, want %q", got, want)
	}
}

func TestPopGameRequiresGameArg(t *testing.T) {
	_, _, err := popGame("run", "<game>", []string{"--blue=./bot"}, []string{"winter2026"})
	if err == nil {
		t.Fatal("popGame returned no error for missing game")
	}
	if !strings.Contains(err.Error(), "usage: arena run <game>") {
		t.Fatalf("error missing usage hint: %v", err)
	}
}

func TestPopGameRejectsUnknownGame(t *testing.T) {
	_, _, err := popGame("run", "<game>", []string{"galactic2099"}, []string{"winter2026"})
	if err == nil {
		t.Fatal("popGame returned no error for unknown game")
	}
	if !strings.Contains(err.Error(), `unknown game "galactic2099"`) {
		t.Fatalf("error missing unknown-game text: %v", err)
	}
}

func TestPopGameUsageIncludesArgsSpec(t *testing.T) {
	_, _, err := popGame("game serialize", "<game> <seed>", nil, []string{"winter2026"})
	if err == nil {
		t.Fatal("popGame returned no error for missing positionals")
	}
	if !strings.Contains(err.Error(), "usage: arena game serialize <game> <seed>") {
		t.Fatalf("error missing full positional spec: %v", err)
	}
}

func TestPrintHelpRoutesReplay(t *testing.T) {
	var stdout bytes.Buffer

	err := printHelp(&stdout, []string{"replay"}, []string{"winter2026"})

	if err != nil {
		t.Fatalf("printHelp returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), "arena replay <game> <username>") {
		t.Fatalf("stdout missing replay usage:\n%s", stdout.String())
	}
}
