package commands

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/craftybase/stocksmith-cli/internal/brand"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print the " + brand.BinaryName + " CLI version, commit, build date, and platform.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s version %s (commit: %s, built: %s, %s, %s/%s)\n",
			brand.BinaryName,
			cliVersion,
			cliCommit,
			cliBuildDate,
			runtime.Version(),
			runtime.GOOS,
			runtime.GOARCH,
		)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var (
	cliVersion   = "dev"
	cliCommit    = "none"
	cliBuildDate = "unknown"
)

func SetVersion(version, commit, buildDate string) {
	cliVersion = version
	cliCommit = commit
	cliBuildDate = buildDate
}
