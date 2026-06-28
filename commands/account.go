package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/craftybase/stocksmith-cli/internal/output"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Show account information",
	Long:  "Display account details including name, currency, time zone, and plan.",
	RunE:  runAccount,
}

func runAccount(cmd *cobra.Command, args []string) error {
	client, _, err := requireAuth()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", client.BaseURL+"/api/v1/account", nil)
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

	if flagJSON {
		return output.PrintJSON(os.Stdout, body)
	}

	var envelope struct {
		Account struct {
			Name     string `json:"name"`
			Currency string `json:"currency_code"`
			TimeZone string `json:"time_zone"`
			Plan     string `json:"plan"`
		} `json:"account"`
	}

	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	a := envelope.Account
	fmt.Fprintf(os.Stdout, "Account:   %s\n", a.Name)
	fmt.Fprintf(os.Stdout, "Currency:  %s\n", a.Currency)
	fmt.Fprintf(os.Stdout, "Time zone: %s\n", a.TimeZone)
	fmt.Fprintf(os.Stdout, "Plan:      %s\n", a.Plan)

	return nil
}

func init() {
	rootCmd.AddCommand(accountCmd)
}
