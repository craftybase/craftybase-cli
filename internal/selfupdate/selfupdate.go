// Package selfupdate replaces the running binary with the latest GitHub release.
// All host/OS/network inputs are injected via Config so the logic is testable.
package selfupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

const (
	defaultAPIBaseURL      = "https://api.github.com"
	defaultDownloadBaseURL = "https://github.com"
)

// Config carries everything the updater needs.
type Config struct {
	BinaryName      string
	Repo            string
	CurrentVersion  string
	GOOS            string
	GOARCH          string
	ExecPath        string
	APIBaseURL      string // default https://api.github.com
	DownloadBaseURL string // default https://github.com
	HTTPClient      *http.Client
	Out             io.Writer
}

func (c *Config) apiBase() string {
	if c.APIBaseURL != "" {
		return c.APIBaseURL
	}
	return defaultAPIBaseURL
}

func (c *Config) downloadBase() string {
	if c.DownloadBaseURL != "" {
		return c.DownloadBaseURL
	}
	return defaultDownloadBaseURL
}

func (c *Config) httpClient() *http.Client {
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	return c.HTTPClient
}

func (c *Config) out() io.Writer {
	if c.Out != nil {
		return c.Out
	}
	return os.Stdout
}

// assetNames returns the release archive and checksums filenames for version,
// matching .goreleaser.yml's name_template.
func (c *Config) assetNames(version string) (archive, checksums string) {
	vnum := strings.TrimPrefix(version, "v")
	archive = fmt.Sprintf("%s_%s_%s_%s.tar.gz", c.BinaryName, vnum, c.GOOS, c.GOARCH)
	checksums = fmt.Sprintf("%s_%s_checksums.txt", c.BinaryName, vnum)
	return archive, checksums
}

// canonicalVersion returns the semver-comparable form ("vX.Y.Z"), adding the
// leading v if absent.
func canonicalVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return v
}

// IsDevVersion reports whether v cannot be compared against released tags.
func IsDevVersion(v string) bool {
	return !semver.IsValid(canonicalVersion(v))
}

// updateAvailable reports whether latest is strictly newer than current.
func updateAvailable(current, latest string) bool {
	return semver.Compare(canonicalVersion(latest), canonicalVersion(current)) > 0
}

// guard returns a refusal error when the running binary must not self-update.
func (c *Config) guard() error {
	if IsDevVersion(c.CurrentVersion) {
		cur := c.CurrentVersion
		if cur == "" {
			cur = "unknown"
		}
		return fmt.Errorf("update is only available for released builds (current version: %s)", cur)
	}
	if c.GOOS == "windows" {
		return fmt.Errorf("self-update isn't supported on Windows — download the latest release zip from https://github.com/%s/releases", c.Repo)
	}
	real, err := filepath.EvalSymlinks(c.ExecPath)
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}
	if isBrewPath(real) {
		return fmt.Errorf("%s was installed via Homebrew — run 'brew upgrade %s' instead", c.BinaryName, c.BinaryName)
	}
	if err := checkWritable(filepath.Dir(real)); err != nil {
		return err
	}
	return nil
}

// isBrewPath reports whether a resolved executable path lives in a Homebrew Cellar.
func isBrewPath(realPath string) bool {
	return strings.Contains(filepath.ToSlash(realPath), "/Cellar/")
}

// checkWritable verifies dir can be written by creating and removing a temp file.
// This is a best-effort pre-flight probe; the real safety guarantee comes from the
// atomic rename in replaceExecutable, so the TOCTOU window here is acceptable.
func checkWritable(dir string) error {
	f, err := os.CreateTemp(dir, ".update-write-test-*")
	if err != nil {
		return fmt.Errorf("%s is not writable — re-run from a shell with permission to write there, or reinstall", dir)
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return nil
}

// LatestVersion returns the newest release tag (e.g. "v0.3.0").
func (c *Config) LatestVersion() (string, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", c.apiBase(), c.Repo)
	body, err := c.get(url)
	if err != nil {
		return "", fmt.Errorf("fetch latest release: %w", err)
	}
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &rel); err != nil {
		return "", fmt.Errorf("parse latest release: %w", err)
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("could not determine latest version")
	}
	return rel.TagName, nil
}

// assetURL builds a release download URL.
func (c *Config) assetURL(version, name string) string {
	return fmt.Sprintf("%s/%s/releases/download/%s/%s", c.downloadBase(), c.Repo, version, name)
}

// get fetches url, returning the body and treating non-2xx as an error.
func (c *Config) get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.BinaryName+"/"+c.CurrentVersion)
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// verifyChecksum confirms archive's SHA-256 matches its entry in a GoReleaser
// checksums.txt body (lines of "<hex>  <filename>").
func verifyChecksum(archive, checksums []byte, archiveName string) error {
	var want string
	for _, line := range strings.Split(string(checksums), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == archiveName {
			want = fields[0]
			break
		}
	}
	if want == "" {
		return fmt.Errorf("no checksum found for %s", archiveName)
	}
	sum := sha256.Sum256(archive)
	if got := hex.EncodeToString(sum[:]); got != want {
		return fmt.Errorf("checksum mismatch for %s (want %s, got %s)", archiveName, want, got)
	}
	return nil
}

// replaceExecutable atomically replaces the binary at execPath (symlinks
// resolved) with newBinary, staging via a temp file in the same directory so the
// final os.Rename is atomic. A running binary is safely replaced on Unix.
func replaceExecutable(execPath string, newBinary []byte) error {
	real, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(real), "."+filepath.Base(real)+"-update-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(newBinary); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write new binary: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("chmod new binary: %w", err)
	}
	if err := os.Rename(tmpName, real); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("replace binary: %w", err)
	}
	return nil
}

// Check reports current vs latest and whether an update is available. Read-only;
// applies no guards (informational even for Homebrew/dev builds).
func (c *Config) Check() (current, latest string, available bool, err error) {
	latest, err = c.LatestVersion()
	if err != nil {
		return c.CurrentVersion, "", false, err
	}
	if IsDevVersion(c.CurrentVersion) {
		return c.CurrentVersion, latest, false, nil
	}
	return c.CurrentVersion, latest, updateAvailable(c.CurrentVersion, latest), nil
}

// Run performs a guarded in-place update to the latest release.
func (c *Config) Run() error {
	if err := c.guard(); err != nil {
		return err
	}
	latest, err := c.LatestVersion()
	if err != nil {
		return err
	}
	if !updateAvailable(c.CurrentVersion, latest) {
		fmt.Fprintf(c.out(), "%s is already up to date (%s)\n", c.BinaryName, c.CurrentVersion)
		return nil
	}

	archiveName, checksumsName := c.assetNames(latest)
	fmt.Fprintf(c.out(), "Downloading %s (%s)…\n", archiveName, latest)
	archive, err := c.get(c.assetURL(latest, archiveName))
	if err != nil {
		return fmt.Errorf("download release: %w", err)
	}
	checksums, err := c.get(c.assetURL(latest, checksumsName))
	if err != nil {
		return fmt.Errorf("download checksums: %w", err)
	}
	if err := verifyChecksum(archive, checksums, archiveName); err != nil {
		return err
	}
	bin, err := extractBinary(archive, c.BinaryName)
	if err != nil {
		return err
	}
	if err := replaceExecutable(c.ExecPath, bin); err != nil {
		return err
	}
	fmt.Fprintf(c.out(), "Updated %s %s → %s\n", c.BinaryName, c.CurrentVersion, strings.TrimPrefix(latest, "v"))
	return nil
}

// extractBinary returns the bytes of the entry named binaryName from a gzipped tar.
func extractBinary(archive []byte, binaryName string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("gunzip archive: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read archive: %w", err)
		}
		if hdr.Typeflag == tar.TypeReg && filepath.Base(hdr.Name) == binaryName {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("read %s from archive: %w", binaryName, err)
			}
			return data, nil
		}
	}
	return nil, fmt.Errorf("archive did not contain %s", binaryName)
}
