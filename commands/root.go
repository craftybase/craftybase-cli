package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/craftybase/stocksmith-cli/internal/api"
	"github.com/craftybase/stocksmith-cli/internal/brand"
	"github.com/craftybase/stocksmith-cli/internal/config"
	"github.com/craftybase/stocksmith-cli/internal/output"
)

var (
	flagToken   string
	flagAPIURL  string
	flagJSON    bool
	flagNDJSON  bool
	flagNoColor bool
	flagVerbose bool
)

var rootCmd = &cobra.Command{
	Use:   brand.BinaryName,
	Short: "Official CLI for the " + brand.ProductName + " Public API",
	Long: brand.BinaryName + ` is a command-line interface for the ` + brand.ProductName + ` Public API.

Authenticate once, then manage your inventory from the terminal.

Documentation: ` + brand.DocsURL,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// RootCmd returns the assembled root command with all subcommands registered.
// Exposed for documentation generation (see cmd/gen-docs).
func RootCmd() *cobra.Command {
	return rootCmd
}

func Execute(version string) {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		if apiErr, ok := err.(*api.APIError); ok {
			os.Exit(api.ExitCode(apiErr))
		}
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "API token (overrides stored credentials)")
	rootCmd.PersistentFlags().StringVar(&flagAPIURL, "api-url", "", "API base URL (default: "+brand.DefaultAPIURL+")")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output raw API envelope (pretty-printed JSON)")
	rootCmd.PersistentFlags().BoolVar(&flagNDJSON, "ndjson", false, "Output auto-paginated NDJSON stream")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable ANSI color output")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Show HTTP request/response detail (token redacted)")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if flagJSON && flagNDJSON {
			return fmt.Errorf("--json and --ndjson are mutually exclusive")
		}
		return nil
	}

	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		if c == rootCmd {
			renderRootHelp(c, c.OutOrStdout(), resolveRenderOpts())
			return
		}
		defaultHelp(c, args)
	})
	rootCmd.Run = func(c *cobra.Command, args []string) {
		renderRootHelp(c, c.OutOrStdout(), resolveRenderOpts())
	}
}

func resolveClient() (*api.Client, *config.Profile, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}

	profile := cfg.ActiveProfile()
	token := config.ResolvedToken(flagToken, profile)
	apiURL := config.ResolvedAPIURL(flagAPIURL, profile)

	client := api.NewClient(apiURL, token, "")
	client.Verbose = flagVerbose

	return client, profile, nil
}

func requireAuth() (*api.Client, *config.Profile, error) {
	client, profile, err := resolveClient()
	if err != nil {
		return nil, nil, err
	}
	if client.Token == "" {
		return nil, nil, &api.APIError{
			StatusCode: 401,
			Message:    "Not authenticated — run '" + brand.BinaryName + " auth login'.",
		}
	}
	return client, profile, nil
}

func resolveRenderOpts() renderOpts {
	color := output.ColorEnabled(flagNoColor)
	width := 0
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}
	return renderOpts{
		color:     color,
		trueColor: color && output.SupportsTrueColor(),
		width:     width,
	}
}
