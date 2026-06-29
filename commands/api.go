package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/craftybase/stocksmith-cli/internal/brand"
	"github.com/craftybase/stocksmith-cli/internal/output"
)

var apiCmd = &cobra.Command{
	Use:   "api <METHOD> <path>",
	Short: "Make authenticated API requests",
	Long: "Make authenticated requests to the " + brand.ProductName + " API.\n\n" +
		"The path must be the full API path starting with /api/v1/.\n\n" +
		"Examples:\n" +
		"  " + brand.BinaryName + " api GET /api/v1/account\n" +
		"  " + brand.BinaryName + " api GET /api/v1/materials\n" +
		"  " + brand.BinaryName + ` api GET "/api/v1/materials?sku=WAX-001"`,
	Args: cobra.ExactArgs(2),
	RunE: runAPI,
}

func runAPI(cmd *cobra.Command, args []string) error {
	client, _, err := requireAuth()
	if err != nil {
		return err
	}

	method := strings.ToUpper(args[0])
	path := args[1]

	if !strings.HasPrefix(path, "/api/v1/") && !strings.HasPrefix(path, "/api/v1") {
		fmt.Fprintf(os.Stderr, "Warning: path should start with /api/v1/ (got %q)\n", path)
	}

	reqURL := client.BaseURL + path
	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	return output.PrintJSON(os.Stdout, body)
}

func init() {
	rootCmd.AddCommand(apiCmd)
}
