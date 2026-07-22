package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	fbcontract "github.com/yuWorm/fba-go/cmd/fbago/internal/contract"
	fbmodule "github.com/yuWorm/fba-go/cmd/fbago/internal/modulecmd"
	fbplugin "github.com/yuWorm/fba-go/cmd/fbago/internal/plugin"
	"github.com/yuWorm/fba-go/cmd/fbago/internal/scaffold"
	fbsecret "github.com/yuWorm/fba-go/cmd/fbago/internal/secret"
	fbswagger "github.com/yuWorm/fba-go/cmd/fbago/internal/swagger"
)

var stdout io.Writer = os.Stdout

func main() {
	if err := execute(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}

func run(args []string) error {
	return execute(args, stdout, io.Discard)
}

func execute(args []string, out, errOut io.Writer) error {
	root := newRootCommand()
	root.SetArgs(args)
	root.SetOut(out)
	root.SetErr(errOut)
	return root.Execute()
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "fbago",
		Short: "Develop and maintain FBA Go projects",
		Long: "fbago creates and maintains FBA Go projects, templates, plugin wiring, " +
			"local module overrides, OpenAPI documents, and API contracts.",
	}
	root.AddCommand(
		newInitCommand(),
		newTemplateCommand(),
		newPluginCommand(),
		newModuleCommand(),
		newSecretCommand(),
		newSwaggerCommand(),
		newContractCommand(),
	)
	return root
}

// Cobra validates flags and arguments before RunE. Silence usage only after that
// validation, so input mistakes include guidance while runtime failures stay concise.
func runAction(cmd *cobra.Command, action func() error) error {
	cmd.Root().SilenceUsage = true
	return action()
}

func newInitCommand() *cobra.Command {
	opts := scaffold.InitOptions{Dir: "."}
	cmd := &cobra.Command{
		Use:   "init <module>",
		Short: "Create a new FBA Go project from a template",
		Args:  cobra.ExactArgs(1),
		Example: "  fbago init github.com/acme/backend --dir ./backend\n" +
			"  fbago init github.com/acme/backend --template basic --dir ./backend",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Module = args[0]
			return runAction(cmd, func() error {
				return scaffold.Init(opts)
			})
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&opts.Dir, "dir", ".", "destination directory")
	flags.BoolVar(&opts.Force, "force", false, "overwrite an existing destination")
	flags.StringVar(&opts.Template, "template", "", "template name, local path, or Git spec (default admin)")
	flags.StringVar(&opts.TemplateReplace, "template-replace", "", "local checkout replacing the template module")
	flags.StringVar(&opts.CoreReplace, "core-replace", "", "local checkout replacing the FBA Go core module")
	flags.StringVar(&opts.CoreVersion, "core-version", "", "FBA Go core version or latest")
	return cmd
}

func newTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Inspect and update template-managed files",
	}
	cmd.AddCommand(
		newTemplateListCommand(),
		newTemplateDiffCommand(),
		newTemplateUpdateCommand(),
	)
	return cmd
}

func newTemplateListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List bundled project templates",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAction(cmd, func() error {
				templates, err := scaffold.ListTemplates()
				if err != nil {
					return err
				}
				for _, template := range templates {
					if _, err := fmt.Fprintln(cmd.OutOrStdout(), template); err != nil {
						return err
					}
				}
				return nil
			})
		},
	}
}

func newTemplateDiffCommand() *cobra.Command {
	var dir string
	var template string
	cmd := &cobra.Command{
		Use:     "diff",
		Short:   "Show changes to template-managed files",
		Args:    cobra.NoArgs,
		Example: "  fbago template diff --dir ./backend",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAction(cmd, func() error {
				result, err := scaffold.DiffTemplate(scaffold.TemplateDiffOptions{
					Dir:      dir,
					Template: template,
				})
				if err != nil {
					return err
				}
				printTemplateChanges(cmd.OutOrStdout(), result.Entries)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&dir, "dir", ".", "project directory")
	cmd.Flags().StringVar(&template, "template", "", "template source override")
	return cmd
}

func newTemplateUpdateCommand() *cobra.Command {
	var dir string
	var template string
	var dryRun bool
	var force bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update template-managed files",
		Args:  cobra.NoArgs,
		Example: "  fbago template update --dry-run\n" +
			"  fbago template update --force",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAction(cmd, func() error {
				result, err := scaffold.UpdateTemplate(scaffold.TemplateUpdateOptions{
					Dir:      dir,
					Template: template,
					DryRun:   dryRun,
					Force:    force,
				})
				if err != nil && len(result.Entries) == 0 {
					return err
				}
				printTemplateChanges(cmd.OutOrStdout(), result.Entries)
				return err
			})
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&dir, "dir", ".", "project directory")
	flags.StringVar(&template, "template", "", "template source override")
	flags.BoolVar(&dryRun, "dry-run", false, "show changes without writing")
	flags.BoolVar(&force, "force", false, "overwrite modified managed files")
	return cmd
}

func printTemplateChanges(out io.Writer, entries []scaffold.TemplateChange) {
	if len(entries) == 0 {
		fmt.Fprintln(out, "no template changes")
		return
	}
	for _, entry := range entries {
		fmt.Fprintf(out, "%s %s\n", entry.Status, entry.Path)
	}
}

func newPluginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Generate and update project plugin wiring",
	}
	cmd.AddCommand(
		newPluginScanCommand(),
		newPluginSyncCommand(),
		newPluginOutdatedCommand(),
		newPluginUpdateCommand(),
	)
	return cmd
}

func newPluginScanCommand() *cobra.Command {
	var mode string
	var moduleDir string
	var pluginsDir string
	var manifest string
	var out string
	var lockOut string
	var packageName string
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan plugin sources and generate registration code",
		Args:  cobra.NoArgs,
		Example: "  fbago plugin scan --mode manifest --manifest plugins.yaml\n" +
			"  fbago plugin scan --mode imports,local --plugins-dir ./plugins",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAction(cmd, func() error {
				result, err := fbplugin.Scan(fbplugin.ScanOptions{
					Modes:      splitModes(mode),
					ModuleDir:  moduleDir,
					PluginsDir: pluginsDir,
					Manifest:   manifest,
				})
				if err != nil {
					return err
				}
				lockPath := lockOut
				if lockPath == "" {
					lockPath = filepath.Join(filepath.Dir(out), "plugin_manifest.lock")
				}
				if err := fbplugin.WriteLock(lockPath, result); err != nil {
					return err
				}
				return fbplugin.GenerateRegistration(out, packageName, result)
			})
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&mode, "mode", "manifest", "comma-separated scan modes")
	flags.StringVar(&moduleDir, "module", ".", "module directory")
	flags.StringVar(&pluginsDir, "plugins-dir", "", "local plugins directory")
	flags.StringVar(&manifest, "manifest", "", "plugins manifest")
	flags.StringVar(&out, "out", "internal/generated/fba_plugins.gen.go", "generated registration output")
	flags.StringVar(&lockOut, "lock-out", "", "plugin lock output")
	flags.StringVar(&packageName, "package", "generated", "generated package name")
	return cmd
}

func newPluginSyncCommand() *cobra.Command {
	var moduleDir string
	var manifest string
	var out string
	var lockOut string
	var packageName string
	var check bool
	cmd := &cobra.Command{
		Use:     "sync",
		Short:   "Synchronize plugin registration, dependencies, and locks",
		Args:    cobra.NoArgs,
		Example: "  fbago plugin sync\n  fbago plugin sync --check",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAction(cmd, func() error {
				if err := fbplugin.Sync(fbplugin.SyncOptions{
					ModuleDir: moduleDir,
					Manifest:  manifest,
					Out:       out,
					LockOut:   lockOut,
					Package:   packageName,
					Check:     check,
				}); err != nil {
					return err
				}
				if check {
					_, err := fmt.Fprintln(cmd.OutOrStdout(), "plugin state is synchronized; dependency updates were not checked")
					return err
				}
				return nil
			})
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&moduleDir, "dir", ".", "project module directory")
	flags.StringVar(&manifest, "manifest", "plugins.yaml", "project plugin manifest")
	flags.StringVar(&out, "out", "internal/generated/fba_plugins.gen.go", "generated registration output")
	flags.StringVar(&lockOut, "lock-out", "plugins.lock", "module-aware plugin lock output")
	flags.StringVar(&packageName, "package", "generated", "generated package name")
	flags.BoolVar(&check, "check", false, "verify generated outputs without writing")
	return cmd
}

func newPluginOutdatedCommand() *cobra.Command {
	var moduleDir string
	var manifest string
	cmd := &cobra.Command{
		Use:   "outdated [plugin-or-module ...]",
		Short: "List available plugin module updates",
		Example: "  fbago plugin outdated\n" +
			"  fbago plugin outdated github.com/acme/fba-plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAction(cmd, func() error {
				statuses, err := fbplugin.Outdated(fbplugin.VersionOptions{
					ModuleDir: moduleDir,
					Manifest:  manifest,
					Targets:   args,
				})
				if err != nil {
					return err
				}
				if len(statuses) == 0 {
					_, err := fmt.Fprintln(cmd.OutOrStdout(), "no plugin modules found")
					return err
				}
				return writePluginModuleStatuses(cmd.OutOrStdout(), statuses)
			})
		},
	}
	cmd.Flags().StringVar(&moduleDir, "dir", ".", "project module directory")
	cmd.Flags().StringVar(&manifest, "manifest", "plugins.yaml", "project plugin manifest")
	return cmd
}

func newPluginUpdateCommand() *cobra.Command {
	var moduleDir string
	var manifest string
	var out string
	var lockOut string
	var packageName string
	var version string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "update [plugin-or-module ...]",
		Short: "Update plugin module dependencies",
		Example: "  fbago plugin update --dry-run\n" +
			"  fbago plugin update github.com/acme/fba-plugin --to v1.2.3",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAction(cmd, func() error {
				updates, err := fbplugin.Update(fbplugin.UpdateOptions{
					ModuleDir: moduleDir,
					Manifest:  manifest,
					Out:       out,
					LockOut:   lockOut,
					Package:   packageName,
					Targets:   args,
					To:        version,
					DryRun:    dryRun,
				})
				if err != nil {
					return err
				}
				return writePluginUpdates(cmd.OutOrStdout(), updates, dryRun)
			})
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&moduleDir, "dir", ".", "project module directory")
	flags.StringVar(&manifest, "manifest", "plugins.yaml", "project plugin manifest")
	flags.StringVar(&out, "out", "internal/generated/fba_plugins.gen.go", "generated registration output")
	flags.StringVar(&lockOut, "lock-out", "plugins.lock", "module-aware plugin lock output")
	flags.StringVar(&packageName, "package", "generated", "generated package name")
	flags.StringVar(&version, "to", "", "target version for one plugin module")
	flags.BoolVar(&dryRun, "dry-run", false, "print updates without changing the project")
	return cmd
}

func writePluginModuleStatuses(out io.Writer, statuses []fbplugin.ModuleStatus) error {
	writer := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "PLUGINS\tMODULE\tCURRENT\tAVAILABLE\tSTATUS"); err != nil {
		return err
	}
	for _, status := range statuses {
		current := displayedVersion(status.Current)
		available := displayedVersion(status.Available)
		state := "current"
		switch {
		case status.Main:
			current = "main"
			available = "-"
			state = "main"
		case status.Replace != nil:
			available = "-"
			state = "replace=" + displayedModuleReplace(status.Replace)
		case status.Available != "" && status.Available != status.Current:
			state = "update"
		}
		if _, err := fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\t%s\n",
			strings.Join(status.PluginIDs, ","),
			status.Module,
			current,
			available,
			state,
		); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func writePluginUpdates(out io.Writer, updates []fbplugin.ModuleUpdate, dryRun bool) error {
	if len(updates) == 0 {
		_, err := fmt.Fprintln(out, "all selected plugin modules are current")
		return err
	}
	action := "updated"
	if dryRun {
		action = "would update"
	}
	for _, update := range updates {
		if _, err := fmt.Fprintf(
			out,
			"%s %s %s -> %s (plugins: %s)\n",
			action,
			update.Module,
			displayedVersion(update.From),
			displayedVersion(update.To),
			strings.Join(update.PluginIDs, ","),
		); err != nil {
			return err
		}
	}
	return nil
}

func displayedVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "-"
	}
	return version
}

func displayedModuleReplace(replacement *fbplugin.ModuleReplace) string {
	if replacement == nil {
		return "-"
	}
	if replacement.Version != "" {
		return replacement.Path + "@" + replacement.Version
	}
	return replacement.Path
}

func splitModes(value string) []string {
	parts := strings.Split(value, ",")
	modes := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			modes = append(modes, part)
		}
	}
	return modes
}

func newModuleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module",
		Short: "Manage local Go module overrides",
	}
	cmd.AddCommand(newModuleUseCommand(), newModuleResetCommand())
	return cmd
}

func newModuleUseCommand() *cobra.Command {
	var projectDir string
	var localPath string
	cmd := &cobra.Command{
		Use:   "use <module>",
		Short: "Use a local checkout for a project dependency",
		Args:  cobra.ExactArgs(1),
		Example: "  fbago module use --path ../fba-go-admin " +
			"github.com/yuWorm/fba-go-admin",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if strings.TrimSpace(localPath) == "" {
				return fmt.Errorf("--path must not be empty")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAction(cmd, func() error {
				return fbmodule.Use(fbmodule.UseOptions{
					ProjectDir: projectDir,
					Module:     args[0],
					Path:       localPath,
				})
			})
		},
	}
	cmd.Flags().StringVar(&projectDir, "dir", ".", "project module directory")
	cmd.Flags().StringVar(&localPath, "path", "", "local module checkout")
	if err := cmd.MarkFlagRequired("path"); err != nil {
		panic(err)
	}
	return cmd
}

func newModuleResetCommand() *cobra.Command {
	var projectDir string
	cmd := &cobra.Command{
		Use:     "reset <module>",
		Short:   "Remove a local module override",
		Args:    cobra.ExactArgs(1),
		Example: "  fbago module reset github.com/yuWorm/fba-go-admin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAction(cmd, func() error {
				return fbmodule.Reset(projectDir, args[0])
			})
		},
	}
	cmd.Flags().StringVar(&projectDir, "dir", ".", "project module directory")
	return cmd
}

func newSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Generate application secrets",
	}
	cmd.AddCommand(newSecretGenerateCommand())
	return cmd
}

func newSecretGenerateCommand() *cobra.Command {
	var size int
	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generate a cryptographically random base64url secret",
		Args:    cobra.NoArgs,
		Example: "  fbago secret generate\n  fbago secret generate --bytes 64",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAction(cmd, func() error {
				value, err := fbsecret.Generate(size)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(cmd.OutOrStdout(), value)
				return err
			})
		},
	}
	cmd.Flags().IntVar(&size, "bytes", fbsecret.DefaultBytes, "cryptographic random bytes before base64url encoding")
	return cmd
}

func newSwaggerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swagger",
		Short: "Build OpenAPI documents",
	}
	cmd.AddCommand(newSwaggerScanCommand())
	return cmd
}

func newSwaggerScanCommand() *cobra.Command {
	var plugins string
	var out string
	var title string
	var version string
	cmd := &cobra.Command{
		Use:     "scan",
		Short:   "Aggregate plugin OpenAPI fragments",
		Args:    cobra.NoArgs,
		Example: "  fbago swagger scan --out docs/openapi.json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAction(cmd, func() error {
				return fbswagger.Scan(fbswagger.ScanOptions{
					PluginLock: plugins,
					Out:        out,
					Title:      title,
					Version:    version,
				})
			})
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&plugins, "plugins", "internal/generated/plugin_manifest.lock", "plugin manifest lock")
	flags.StringVar(&out, "out", "docs/openapi.json", "OpenAPI output")
	flags.StringVar(&title, "title", "FBA API", "document title")
	flags.StringVar(&version, "version", "0.1.0", "document version")
	return cmd
}

func newContractCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract",
		Short: "Manage API contract snapshots and checks",
	}
	cmd.AddCommand(newContractSnapshotCommand(), newContractTestCommand())
	return cmd
}

func newContractSnapshotCommand() *cobra.Command {
	var contractDir string
	var out string
	cmd := &cobra.Command{
		Use:     "snapshot",
		Short:   "Generate an API contract snapshot",
		Args:    cobra.NoArgs,
		Example: "  fbago contract snapshot --contract contracts --out internal/generated/api.contract.snapshot.json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAction(cmd, func() error {
				contracts, err := fbcontract.Load(contractDir)
				if err != nil {
					return err
				}
				snapshot, err := fbcontract.Snapshot(contracts)
				if err != nil {
					return err
				}
				return fbcontract.WriteSnapshot(out, snapshot)
			})
		},
	}
	cmd.Flags().StringVar(&contractDir, "contract", "contracts", "contract directory")
	cmd.Flags().StringVar(&out, "out", "internal/generated/api.contract.snapshot.json", "snapshot output")
	return cmd
}

func newContractTestCommand() *cobra.Command {
	var baseURL string
	var contractDir string
	cmd := &cobra.Command{
		Use:     "test",
		Short:   "Run API contract checks against a server",
		Args:    cobra.NoArgs,
		Example: "  fbago contract test --base-url http://127.0.0.1:8001",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAction(cmd, func() error {
				contracts, err := fbcontract.Load(contractDir)
				if err != nil {
					return err
				}
				result, err := fbcontract.Test(fbcontract.TestOptions{
					BaseURL:   baseURL,
					Contracts: contracts,
				})
				if err != nil {
					return err
				}
				if !result.Passed {
					return fmt.Errorf("%s", fbcontract.FormatFailures(result))
				}
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&baseURL, "base-url", "", "base URL")
	cmd.Flags().StringVar(&contractDir, "contract", "contracts", "contract directory")
	return cmd
}
