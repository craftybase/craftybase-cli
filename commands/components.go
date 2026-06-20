package commands

import "github.com/spf13/cobra"

var componentsCmd = &cobra.Command{
	Use:   "components",
	Short: "Manage components",
}

var componentsListFlags projectListFlags

func init() {
	res := projectResource{
		pathSegment: "components",
		collection:  "components",
		singular:    "component",
	}
	componentsCmd.AddCommand(newProjectListCmd(res, &componentsListFlags))
	componentsCmd.AddCommand(newProjectShowCmd(res))
	rootCmd.AddCommand(componentsCmd)
}
