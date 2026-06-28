package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/craftybase/stocksmith-cli/internal/api"
	"github.com/craftybase/stocksmith-cli/internal/brand"
	"github.com/craftybase/stocksmith-cli/internal/config"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication credentials",
}

var authLoginToken string

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the " + brand.ProductName + " API",
	Long: "Authenticate with the " + brand.ProductName + " API using an API key.\n\n" +
		"The key can be provided via:\n" +
		"  - --token flag\n" +
		"  - stdin (when piped)\n" +
		"  - interactive prompt (when run in a terminal)\n\n" +
		"On success, credentials are saved to ~/" + brand.ConfigDir + "/" + brand.ConfigFile + ".",
	RunE: runAuthLogin,
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	apiURL := config.ResolvedAPIURL(flagAPIURL, nil)

	var token string

	if authLoginToken != "" {
		token = authLoginToken
	} else if !term.IsTerminal(int(os.Stdin.Fd())) {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read token from stdin: %w", err)
		}
		token = strings.TrimSpace(string(data))
	} else {
		fmt.Fprint(os.Stderr, "Paste your "+brand.ProductName+" API key (input hidden): ")
		raw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}
		token = strings.TrimSpace(string(raw))
	}

	if token == "" {
		return &api.APIError{StatusCode: 401, Message: "No API key provided."}
	}

	client := api.NewClient(apiURL, token, cliVersion)
	client.Verbose = flagVerbose

	req, err := http.NewRequest("GET", apiURL+"/api/v1/account", nil)
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

	var envelope struct {
		Account struct {
			Name string `json:"name"`
			ID   int    `json:"id"`
		} `json:"account"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("parse account response: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	profile := cfg.ActiveProfile()
	profile.Token = token
	profile.AccountName = envelope.Account.Name
	profile.AccountID = envelope.Account.ID

	if apiURL != brand.DefaultAPIURL {
		profile.APIURL = apiURL
	} else {
		profile.APIURL = ""
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save credentials: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Authenticated as %q. Credentials saved to ~/.%s/%s\n",
		envelope.Account.Name, brand.ConfigDir, brand.ConfigFile)
	return nil
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE:  runAuthStatus,
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	profile := cfg.ActiveProfile()
	token, source := config.ResolvedTokenWithSource(flagToken, profile)

	if token == "" {
		return &api.APIError{
			StatusCode: 401,
			Message:    "Not authenticated — run '" + brand.BinaryName + " auth login'.",
		}
	}

	apiURL := config.ResolvedAPIURL(flagAPIURL, profile)
	client := api.NewClient(apiURL, token, cliVersion)
	client.Verbose = flagVerbose

	req, err := http.NewRequest("GET", apiURL+"/api/v1/ping", nil)
	if err != nil {
		return fmt.Errorf("build ping request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read ping response: %w", err)
	}

	var pingResp struct {
		AccountID int `json:"account_id"`
	}
	if err := json.Unmarshal(body, &pingResp); err != nil {
		// Non-fatal: proceed without account_id from ping
		pingResp.AccountID = 0
	}

	accountName := profile.AccountName

	if source == config.TokenSourceEnv || source == config.TokenSourceFlag {
		accountReq, err := http.NewRequest("GET", apiURL+"/api/v1/account", nil)
		if err == nil {
			if accountResp, err := client.Do(accountReq); err == nil {
				defer accountResp.Body.Close()
				accountBody, _ := io.ReadAll(accountResp.Body)
				var accountEnvelope struct {
					Account struct {
						Name string `json:"name"`
					} `json:"account"`
				}
				if json.Unmarshal(accountBody, &accountEnvelope) == nil {
					accountName = accountEnvelope.Account.Name
				}
			}
		}
	}

	if accountName == "" && pingResp.AccountID > 0 {
		accountName = fmt.Sprintf("account:%d", pingResp.AccountID)
	}

	if profile.AccountID > 0 && pingResp.AccountID > 0 && profile.AccountID != pingResp.AccountID {
		fmt.Fprintf(os.Stderr, "Warning: account ID mismatch — re-run '%s auth login' to refresh cached credentials.\n", brand.BinaryName)
	}

	maskedKey := maskToken(token)
	fmt.Fprintf(os.Stdout, "Authenticated   account: %s   key: %s   url: %s\n",
		accountName, maskedKey, apiURL)
	return nil
}

func maskToken(token string) string {
	if len(token) <= 9 {
		return "***"
	}
	last4 := token[len(token)-4:]
	prefix := token
	if len(prefix) > 9 {
		prefix = prefix[:9]
	}
	return prefix + "…" + last4
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	RunE:  runAuthLogout,
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		fmt.Fprintln(os.Stdout, "No stored credentials found.")
		return nil
	}

	if cfg.Profiles == nil || cfg.Profiles["default"] == nil {
		fmt.Fprintln(os.Stdout, "No stored credentials found.")
		return nil
	}

	delete(cfg.Profiles, "default")

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("remove credentials: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Logged out. Credentials removed from ~/.%s/%s\n",
		brand.ConfigDir, brand.ConfigFile)
	return nil
}

func init() {
	authLoginCmd.Flags().StringVar(&authLoginToken, "token", "", "API token to authenticate with")
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}
