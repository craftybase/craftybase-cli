package commands

import "github.com/spf13/cobra"

var productsCmd = &cobra.Command{
	Use:   "products",
	Short: "Manage products",
}

var productsListFlags resourceListFlags

func init() {
	res := resourceConfig{
		pathSegment: "products",
		collection:  "products",
		singular:    "product",
		toTable:     projectsToTable,
		renderShow:  renderProjectShowRaw,
	}
	productsCmd.AddCommand(newResourceListCmd(res, &productsListFlags))
	productsCmd.AddCommand(newResourceShowCmd(res))
	rootCmd.AddCommand(productsCmd)
}
