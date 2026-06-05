package command_test

import (
	"context"
	"strings"
	"testing"

	"github.com/yuWorm/fba-go/core/command"
)

func TestExecuteRunsDefaultCommandWhenNoArgs(t *testing.T) {
	var called bool
	err := command.Execute(context.Background(), command.ExecuteOptions{
		Use:            "admin",
		DefaultCommand: "server",
		Commands: []command.Command{{
			Use:   "server",
			Short: "Start HTTP server",
			Run: func(context.Context, command.Runtime, []string) error {
				called = true
				return nil
			},
		}},
	}, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !called {
		t.Fatal("default command was not called")
	}
}

func TestExecuteRunsNestedCommandWithArgs(t *testing.T) {
	var gotArgs []string
	err := command.Execute(context.Background(), command.ExecuteOptions{
		Use: "admin",
		Commands: []command.Command{{
			Use:                "task worker",
			Short:              "Start task worker",
			DisableFlagParsing: true,
			Run: func(_ context.Context, _ command.Runtime, args []string) error {
				gotArgs = append([]string(nil), args...)
				return nil
			},
		}},
	}, []string{"task", "worker", "--queue", "default"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	got := strings.Join(gotArgs, " ")
	if got != "--queue default" {
		t.Fatalf("args = %q, want --queue default", got)
	}
}

func TestExecuteRejectsDuplicateCommandUse(t *testing.T) {
	err := command.Execute(context.Background(), command.ExecuteOptions{
		Use: "admin",
		Commands: []command.Command{
			{Use: "server", Run: func(context.Context, command.Runtime, []string) error { return nil }},
			{Use: "server", Run: func(context.Context, command.Runtime, []string) error { return nil }},
		},
	}, []string{"server"})
	if err == nil || !strings.Contains(err.Error(), "duplicate command") {
		t.Fatalf("Execute() error = %v, want duplicate command", err)
	}
}
