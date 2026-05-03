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
	if !strings.Contains(stdout.String(), "Usage: arena <command> [OPTIONS]") {
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

func TestNormalizeCommandSupportsImplicitRun(t *testing.T) {
	command, rest := normalizeCommand([]string{"--blue=./bot", "--game=winter2026"})

	if command != "run" {
		t.Fatalf("command = %q, want run", command)
	}
	if got, want := strings.Join(rest, " "), "--blue=./bot --game=winter2026"; got != want {
		t.Fatalf("rest = %q, want %q", got, want)
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
	spec, rest, err := selectCommand("replay", []string{"mrsombre", "123"})
	if err != nil {
		t.Fatalf("selectCommand returned error: %v", err)
	}
	if spec.handler == nil {
		t.Fatal("handler is nil")
	}
	if !spec.needsFactory {
		t.Fatal("needsFactory = false, want true")
	}
	if got, want := strings.Join(rest, " "), "mrsombre 123"; got != want {
		t.Fatalf("rest = %q, want %q", got, want)
	}
}

func TestPrintHelpRoutesReplay(t *testing.T) {
	var stdout bytes.Buffer

	err := printHelp(&stdout, []string{"replay"}, []string{"winter2026"})

	if err != nil {
		t.Fatalf("printHelp returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), "arena replay <username>") {
		t.Fatalf("stdout missing replay usage:\n%s", stdout.String())
	}
}
