package commands

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdate_DevBuildRefuses(t *testing.T) {
	// cliVersion defaults to "dev" in tests, so the guard must refuse.
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"update"})
	t.Cleanup(func() { rootCmd.SetOut(nil); rootCmd.SetArgs(nil); updateCheckOnly = false })

	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "released builds") {
		t.Errorf("dev-build update should refuse, got err=%v", err)
	}
}

func TestUpdate_CheckDevBuild(t *testing.T) {
	// cliVersion defaults to "dev" in tests; --check should report it as a dev build.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v9.9.9"}`))
	}))
	defer srv.Close()

	updateAPIBaseURL = srv.URL
	t.Cleanup(func() {
		updateAPIBaseURL = ""
		updateCheckOnly = false
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
	})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"update", "--check"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "dev build") {
		t.Errorf("--check dev build output missing 'dev build': %q", out)
	}
	if !strings.Contains(out, "v9.9.9") {
		t.Errorf("--check dev build output missing latest version 'v9.9.9': %q", out)
	}
}

func TestUpdate_CheckReportsAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v9.9.9"}`))
	}))
	defer srv.Close()

	SetVersion("0.2.0", "abc", "today")
	updateAPIBaseURL = srv.URL
	t.Cleanup(func() {
		SetVersion("dev", "none", "unknown")
		updateAPIBaseURL = ""
		updateCheckOnly = false
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
	})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"update", "--check"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "craftybase v9.9.9 is available") {
		t.Errorf("--check output = %q", buf.String())
	}
}
