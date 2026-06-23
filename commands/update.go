package commands

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"github.com/craftybase/craftybase-cli/internal/brand"
	"github.com/craftybase/craftybase-cli/internal/selfupdate"
)

// Overridable in tests to point at an httptest server; empty => real GitHub hosts.
var (
	updateAPIBaseURL      = ""
	updateDownloadBaseURL = ""
)

var updateCheckOnly bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update " + brand.BinaryName + " to the latest release",
	Long: `Update the ` + brand.BinaryName + ` binary in place to the latest GitHub release.

Downloads the release archive for your platform, verifies its SHA-256 checksum,
and atomically replaces the running binary. Homebrew installs should use
'brew upgrade ` + brand.BinaryName + `'; Windows users download the release zip manually.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		cfg := selfupdate.Config{
			BinaryName:      brand.BinaryName,
			Repo:            brand.GitHubRepo,
			CurrentVersion:  cliVersion,
			GOOS:            runtime.GOOS,
			GOARCH:          runtime.GOARCH,
			ExecPath:        exe,
			APIBaseURL:      updateAPIBaseURL,
			DownloadBaseURL: updateDownloadBaseURL,
			HTTPClient:      &http.Client{Timeout: 30 * time.Second},
			Out:             cmd.OutOrStdout(),
		}
		if updateCheckOnly {
			current, latest, available, err := cfg.Check()
			if err != nil {
				return err
			}
			switch {
			case selfupdate.IsDevVersion(current):
				fmt.Fprintf(cmd.OutOrStdout(), "%s is a dev build (no released version to compare); latest release is %s\n", brand.BinaryName, latest)
			case available:
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s is available (current: %s). Run '%s update' to upgrade.\n",
					brand.BinaryName, latest, current, brand.BinaryName)
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "%s is up to date (current: %s, latest: %s)\n",
					brand.BinaryName, current, latest)
			}
			return nil
		}
		return cfg.Run()
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheckOnly, "check", false, "Check for a newer release without installing")
	rootCmd.AddCommand(updateCmd)
}
