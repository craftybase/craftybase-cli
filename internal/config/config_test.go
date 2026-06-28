package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/craftybase/stocksmith-cli/internal/brand"
	"github.com/craftybase/stocksmith-cli/internal/config"
)

func TestLoadMissingFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := &config.Config{}
	p := cfg.ActiveProfile()
	p.Token = "live_testtoken123"
	p.AccountName = "TestAccount"
	p.AccountID = 42
	p.APIURL = brand.DefaultAPIURL

	if err := config.Save(cfg); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	lp := loaded.ActiveProfile()
	if lp.Token != "live_testtoken123" {
		t.Errorf("token mismatch: got %q", lp.Token)
	}
	if lp.AccountName != "TestAccount" {
		t.Errorf("account_name mismatch: got %q", lp.AccountName)
	}
	if lp.AccountID != 42 {
		t.Errorf("account_id mismatch: got %d", lp.AccountID)
	}
}

func TestLoad_PermissionWarning(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, brand.ConfigDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, brand.ConfigFile)

	content := "[profiles.default]\ntoken = \"live_abc\"\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load should succeed even with loose permissions, got: %v", err)
	}
	if cfg.ActiveProfile().Token != "live_abc" {
		t.Errorf("expected token live_abc, got %q", cfg.ActiveProfile().Token)
	}
}

func TestResolvedToken_Precedence(t *testing.T) {
	t.Setenv("STOCKSMITH_API_TOKEN", "")

	profile := &config.Profile{Token: "profile_token"}

	tok := config.ResolvedToken("flag_token", profile)
	if tok != "flag_token" {
		t.Errorf("expected flag_token, got %q", tok)
	}

	t.Setenv("STOCKSMITH_API_TOKEN", "env_token")
	tok = config.ResolvedToken("", profile)
	if tok != "env_token" {
		t.Errorf("expected env_token, got %q", tok)
	}

	t.Setenv("STOCKSMITH_API_TOKEN", "")
	tok = config.ResolvedToken("", profile)
	if tok != "profile_token" {
		t.Errorf("expected profile_token, got %q", tok)
	}

	tok = config.ResolvedToken("", nil)
	if tok != "" {
		t.Errorf("expected empty token, got %q", tok)
	}
}

func TestResolvedAPIURL_Precedence(t *testing.T) {
	t.Setenv("STOCKSMITH_API_URL", "")

	profile := &config.Profile{APIURL: "https://staging.stocksmith.com"}

	url := config.ResolvedAPIURL("https://flag.stocksmith.com", profile)
	if url != "https://flag.stocksmith.com" {
		t.Errorf("expected flag URL, got %q", url)
	}

	t.Setenv("STOCKSMITH_API_URL", "https://env.stocksmith.com")
	url = config.ResolvedAPIURL("", profile)
	if url != "https://env.stocksmith.com" {
		t.Errorf("expected env URL, got %q", url)
	}

	t.Setenv("STOCKSMITH_API_URL", "")
	url = config.ResolvedAPIURL("", profile)
	if url != "https://staging.stocksmith.com" {
		t.Errorf("expected profile URL, got %q", url)
	}

	url = config.ResolvedAPIURL("", nil)
	if url != brand.DefaultAPIURL {
		t.Errorf("expected default URL, got %q", url)
	}
}
