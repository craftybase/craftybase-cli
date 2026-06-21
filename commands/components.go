package commands

import "github.com/spf13/cobra"

var componentsCmd = &cobra.Command{
	Use:   "components",
	Short: "Manage components",
}

var (
	componentsFilters    projectFilters
	componentsPagination paginationFlags
)

func init() {
	res := resourceConfig{
		pathSegment: "components",
		collection:  "components",
		singular:    "component",
		toTable:     projectsToTable,
		renderShow:  renderProjectShowRaw,
	}
	componentsCmd.AddCommand(newResourceListCmd(res, &componentsFilters, &componentsPagination))
	componentsCmd.AddCommand(newResourceShowCmd(res))
	rootCmd.AddCommand(componentsCmd)
}
