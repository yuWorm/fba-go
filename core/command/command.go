package command

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/di"
)

type Runtime interface {
	Container() *di.Container
	Config() config.Options
	Output() io.Writer
	ErrorOutput() io.Writer
}

type Handler func(context.Context, Runtime, []string) error

type Command struct {
	Use                string
	Short              string
	Long               string
	Aliases            []string
	DisableFlagParsing bool
	Run                Handler
}

type ExecuteOptions struct {
	Use            string
	Short          string
	Runtime        Runtime
	Commands       []Command
	DefaultCommand string
	Out            io.Writer
	Err            io.Writer
}

func Execute(ctx context.Context, opts ExecuteOptions, args []string) error {
	root, err := NewRoot(opts)
	if err != nil {
		return err
	}
	root.SetArgs(args)
	return root.ExecuteContext(ctx)
}

func NewRoot(opts ExecuteOptions) (*cobra.Command, error) {
	use := strings.TrimSpace(opts.Use)
	if use == "" {
		use = "app"
	}
	root := &cobra.Command{
		Use:           use,
		Short:         opts.Short,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("unknown command %s", strings.Join(args, " "))
			}
			if opts.DefaultCommand == "" {
				return cmd.Help()
			}
			command, ok := findCommand(opts.Commands, opts.DefaultCommand)
			if !ok {
				return fmt.Errorf("default command %q is not registered", opts.DefaultCommand)
			}
			if command.Run == nil {
				return fmt.Errorf("default command %q has no handler", opts.DefaultCommand)
			}
			return command.Run(cmd.Context(), opts.Runtime, nil)
		},
	}
	if opts.Out != nil {
		root.SetOut(opts.Out)
	}
	if opts.Err != nil {
		root.SetErr(opts.Err)
	}

	seen := make(map[string]bool, len(opts.Commands))
	nodes := map[string]*cobra.Command{"": root}
	for _, command := range opts.Commands {
		if err := installCommand(root, nodes, seen, command, opts.Runtime); err != nil {
			return nil, err
		}
	}
	return root, nil
}

func installCommand(root *cobra.Command, nodes map[string]*cobra.Command, seen map[string]bool, command Command, runtime Runtime) error {
	path := normalizeUse(command.Use)
	if path == "" {
		return fmt.Errorf("command use is required")
	}
	if seen[path] {
		return fmt.Errorf("duplicate command %q", path)
	}
	seen[path] = true

	parts := strings.Split(path, " ")
	parent := root
	parentPath := ""
	for i, part := range parts {
		currentPath := strings.TrimSpace(parentPath + " " + part)
		node, ok := nodes[currentPath]
		if !ok {
			node = &cobra.Command{
				Use:           part,
				SilenceUsage:  true,
				SilenceErrors: true,
			}
			nodes[currentPath] = node
			parent.AddCommand(node)
		}
		if i == len(parts)-1 {
			applyCommand(node, command, runtime)
		}
		parent = node
		parentPath = currentPath
	}
	return nil
}

func applyCommand(node *cobra.Command, command Command, runtime Runtime) {
	node.Short = command.Short
	node.Long = command.Long
	node.Aliases = append([]string(nil), command.Aliases...)
	// Some plugin commands need raw passthrough args for their own parsers, but
	// Cobra flag parsing stays enabled by default so help and flags keep working.
	node.DisableFlagParsing = command.DisableFlagParsing
	node.RunE = func(cmd *cobra.Command, args []string) error {
		if command.Run == nil {
			return cmd.Help()
		}
		return command.Run(cmd.Context(), runtime, args)
	}
}

func findCommand(commands []Command, use string) (Command, bool) {
	needle := normalizeUse(use)
	for _, command := range commands {
		if normalizeUse(command.Use) == needle {
			return command, true
		}
	}
	return Command{}, false
}

func normalizeUse(use string) string {
	return strings.Join(strings.Fields(use), " ")
}
