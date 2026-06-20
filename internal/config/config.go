package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"github.com/craftybase/craftybase-cli/internal/brand"
)

type Profile struct {
	Token       string `toml:"token"`
	APIURL      string `toml:"api_url,omitempty"`
	AccountName string `toml:"account_name,omitempty"`
	AccountID   int    `toml:"account_id,omitempty"`
}

type Config struct {
	Profiles map[string]*Profile `toml:"profiles"`
}

func (c *Config) ActiveProfile() *Profile {
	if c.Profiles == nil {
		c.Profiles = make(map[string]*Profile)
	}
	p, ok := c.Profiles["default"]
	if !ok {
		p = &Profile{}
		c.Profiles["default"] = p
	}
	return p
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, brand.ConfigDir, brand.ConfigFile), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return &Config{}, err
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return &Config{}, fmt.Errorf("stat config file: %w", err)
	}

	mode := info.Mode().Perm()
	if mode != 0o600 {
		fmt.Fprintf(os.Stderr, "Warning: credentials file %s has loose permissions (%s). Run: chmod 0600 %s\n",
			path, mode, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{}, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return &Config{}, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("set config directory permissions: %w", err)
	}

	tmpPath := path + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		os.Remove(tmpPath)
		f, err = os.OpenFile(tmpPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("create temp config file: %w", err)
		}
	}

	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("encode config: %w", err)
	}

	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp config file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename config file: %w", err)
	}

	return nil
}

func ResolvedToken(flagToken string, profile *Profile) string {
	if flagToken != "" {
		return flagToken
	}
	if v := os.Getenv(brand.EnvTokenName); v != "" {
		return v
	}
	if profile != nil && profile.Token != "" {
		return profile.Token
	}
	return ""
}

type TokenSource int

const (
	TokenSourceNone TokenSource = iota
	TokenSourceFlag
	TokenSourceEnv
	TokenSourceProfile
)

func ResolvedTokenWithSource(flagToken string, profile *Profile) (string, TokenSource) {
	if flagToken != "" {
		return flagToken, TokenSourceFlag
	}
	if v := os.Getenv(brand.EnvTokenName); v != "" {
		return v, TokenSourceEnv
	}
	if profile != nil && profile.Token != "" {
		return profile.Token, TokenSourceProfile
	}
	return "", TokenSourceNone
}

func ResolvedAPIURL(flagURL string, profile *Profile) string {
	if flagURL != "" {
		return flagURL
	}
	if v := os.Getenv(brand.EnvAPIURL); v != "" {
		return v
	}
	if profile != nil && profile.APIURL != "" {
		return profile.APIURL
	}
	return brand.DefaultAPIURL
}
