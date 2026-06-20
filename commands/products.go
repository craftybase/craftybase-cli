package commands

import "github.com/spf13/cobra"

var productsCmd = &cobra.Command{
	Use:   "products",
	Short: "Manage products",
}

var productsListFlags projectListFlags

func init() {
	res := projectResource{
		pathSegment: "products",
		collection:  "products",
		singular:    "product",
	}
	productsCmd.AddCommand(newProjectListCmd(res, &productsListFlags))
	productsCmd.AddCommand(newProjectShowCmd(res))
	rootCmd.AddCommand(productsCmd)
}
