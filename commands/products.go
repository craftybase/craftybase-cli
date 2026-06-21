package commands

import "github.com/spf13/cobra"

var productsCmd = &cobra.Command{
	Use:   "products",
	Short: "Manage products",
}

var (
	productsFilters    projectFilters
	productsPagination paginationFlags
)

func init() {
	res := resourceConfig{
		pathSegment: "products",
		collection:  "products",
		singular:    "product",
		toTable:     projectsToTable,
		renderShow:  renderProjectShowRaw,
	}
	productsCmd.AddCommand(newResourceListCmd(res, &productsFilters, &productsPagination))
	productsCmd.AddCommand(newResourceShowCmd(res))
	rootCmd.AddCommand(productsCmd)
}
