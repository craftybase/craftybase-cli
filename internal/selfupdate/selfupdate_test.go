package selfupdate

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssetNames(t *testing.T) {
	c := Config{BinaryName: "craftybase", GOOS: "darwin", GOARCH: "arm64"}
	archive, checksums := c.assetNames("v0.3.0")
	if archive != "craftybase_0.3.0_darwin_arm64.tar.gz" {
		t.Errorf("archive = %q", archive)
	}
	if checksums != "craftybase_0.3.0_checksums.txt" {
		t.Errorf("checksums = %q", checksums)
	}
}

func TestIsDevVersion(t *testing.T) {
	for _, v := range []string{"", "dev", "garbage", "none"} {
		if !isDevVersion(v) {
			t.Errorf("isDevVersion(%q) = false, want true", v)
		}
	}
	for _, v := range []string{"0.2.0", "v0.2.0", "1.0.0-rc1"} {
		if isDevVersion(v) {
			t.Errorf("isDevVersion(%q) = true, want false", v)
		}
	}
}

func TestUpdateAvailable(t *testing.T) {
	if !updateAvailable("0.2.0", "v0.3.0") {
		t.Error("0.2.0 -> v0.3.0 should be available")
	}
	if updateAvailable("0.3.0", "v0.3.0") {
		t.Error("equal versions: not available")
	}
	if updateAvailable("0.4.0", "v0.3.0") {
		t.Error("current ahead of latest: not available")
	}
}

func TestIsBrewPath(t *testing.T) {
	if !isBrewPath("/opt/homebrew/Cellar/craftybase/0.2.0/bin/craftybase") {
		t.Error("brew Cellar path should be detected")
	}
	if isBrewPath("/Users/x/.local/bin/craftybase") {
		t.Error(".local/bin path is not brew")
	}
}

func TestGuardDevRefuses(t *testing.T) {
	c := Config{BinaryName: "craftybase", CurrentVersion: "dev", GOOS: "darwin"}
	err := c.guard()
	if err == nil || !strings.Contains(err.Error(), "released builds") {
		t.Errorf("dev guard err = %v", err)
	}
}

func TestGuardWindowsRefuses(t *testing.T) {
	c := Config{BinaryName: "craftybase", CurrentVersion: "0.2.0", GOOS: "windows", Repo: "craftybase/craftybase-cli"}
	err := c.guard()
	if err == nil || !strings.Contains(err.Error(), "Windows") {
		t.Errorf("windows guard err = %v", err)
	}
}

func TestGuardBrewRefuses(t *testing.T) {
	// Build a real Cellar-shaped path with a symlink, since guard() EvalSymlinks.
	root := t.TempDir()
	cellarBin := filepath.Join(root, "Cellar", "craftybase", "0.2.0", "bin")
	if err := os.MkdirAll(cellarBin, 0o755); err != nil {
		t.Fatal(err)
	}
	realBin := filepath.Join(cellarBin, "craftybase")
	if err := os.WriteFile(realBin, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(linkDir, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(linkDir, "craftybase")
	if err := os.Symlink(realBin, link); err != nil {
		t.Fatal(err)
	}
	c := Config{BinaryName: "craftybase", CurrentVersion: "0.2.0", GOOS: "darwin", ExecPath: link}
	err := c.guard()
	if err == nil || !strings.Contains(err.Error(), "Homebrew") {
		t.Errorf("brew guard err = %v", err)
	}
}

func TestGuardPassesForWritableLocalBin(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "craftybase")
	if err := os.WriteFile(exe, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	c := Config{BinaryName: "craftybase", CurrentVersion: "0.2.0", GOOS: "darwin", ExecPath: exe}
	if err := c.guard(); err != nil {
		t.Errorf("guard should pass, got %v", err)
	}
}

func TestLatestVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/craftybase/craftybase-cli/releases/latest" {
			_, _ = w.Write([]byte(`{"tag_name":"v0.3.0"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := Config{Repo: "craftybase/craftybase-cli", APIBaseURL: srv.URL}
	got, err := c.LatestVersion()
	if err != nil {
		t.Fatal(err)
	}
	if got != "v0.3.0" {
		t.Errorf("LatestVersion = %q, want v0.3.0", got)
	}
}
