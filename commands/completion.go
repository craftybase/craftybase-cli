package commands

import (
	"github.com/spf13/cobra"

	"github.com/craftybase/stocksmith-cli/internal/brand"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: "Generate shell completion scripts for " + brand.BinaryName + ".\n\n" +
		"To load completions:\n\n" +
		"Bash:\n" +
		"  $ source <(" + brand.BinaryName + " completion bash)\n\n" +
		"Zsh:\n" +
		"  $ source <(" + brand.BinaryName + " completion zsh)\n\n" +
		"Fish:\n" +
		"  $ " + brand.BinaryName + " completion fish | source\n\n" +
		"PowerShell:\n" +
		"  PS> " + brand.BinaryName + " completion powershell | Out-String | Invoke-Expression\n",
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(cmd.OutOrStdout())
		case "zsh":
			return rootCmd.GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
