package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/craftybase/craftybase-cli/internal/api"
)

func newTestClient(server *httptest.Server) *api.Client {
	c := api.NewClient(server.URL, "test_token", "1.0.0-test")
	return c
}

func TestClient_AuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	req, _ := http.NewRequest("GET", srv.URL+"/api/v1/materials", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if gotAuth != "Bearer test_token" {
		t.Errorf("expected 'Bearer test_token', got %q", gotAuth)
	}
}

func TestClient_RequestPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	req, _ := http.NewRequest("GET", srv.URL+"/api/v1/materials", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if gotPath != "/api/v1/materials" {
		t.Errorf("expected path /api/v1/materials, got %q", gotPath)
	}
}

func TestClient_UserAgent(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	req, _ := http.NewRequest("GET", srv.URL+"/api/v1/account", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if !strings.HasPrefix(gotUA, "craftybase-cli/") {
		t.Errorf("User-Agent should start with 'craftybase-cli/', got %q", gotUA)
	}
}

func TestClient_Retry429_HonorsRetryAfter(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	req, _ := http.NewRequest("GET", srv.URL+"/api/v1/account", nil)
	start := time.Now()
	resp, err := c.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	resp.Body.Close()

	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
	if elapsed < 900*time.Millisecond {
		t.Errorf("expected to wait ~1s for Retry-After, only waited %s", elapsed)
	}
}

func TestClient_Retry429_NoRetryAfterFallback(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = atomic.AddInt32(&callCount, 1)
		w.WriteHeader(429)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	c.HTTPClient = &http.Client{Timeout: 5 * time.Second}

	req, _ := http.NewRequest("GET", srv.URL+"/api/v1/account", nil)
	_, err := c.Do(req)
	if err == nil {
		t.Fatal("expected error from exhausted 429 retries")
	}
	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 429 {
		t.Errorf("expected status 429, got %d", apiErr.StatusCode)
	}
}

func TestClient_ErrorMapping_401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"invalid token"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	req, _ := http.NewRequest("GET", srv.URL+"/api/v1/account", nil)
	_, err := c.Do(req)

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected 401, got %d", apiErr.StatusCode)
	}
	if api.ExitCode(err) != 3 {
		t.Errorf("expected exit code 3 for 401, got %d", api.ExitCode(err))
	}
}

func TestClient_ErrorMapping_403(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte(`{"error":"API access not enabled"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	req, _ := http.NewRequest("GET", srv.URL+"/api/v1/account", nil)
	_, err := c.Do(req)

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("expected 403, got %d", apiErr.StatusCode)
	}
	if api.ExitCode(err) != 1 {
		t.Errorf("expected exit code 1 for 403, got %d", api.ExitCode(err))
	}
	if apiErr.Message != "API access not enabled" {
		t.Errorf("expected API body message, got %q", apiErr.Message)
	}
}

func TestClient_ErrorMapping_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"error":"Material not found."}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	req, _ := http.NewRequest("GET", srv.URL+"/api/v1/materials/999", nil)
	_, err := c.Do(req)

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
	if api.ExitCode(err) != 4 {
		t.Errorf("expected exit code 4 for 404, got %d", api.ExitCode(err))
	}
}

func TestClient_ErrorMapping_5xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-Id", "req-123")
		w.WriteHeader(503)
		w.Write([]byte(`{"error":"Service temporarily unavailable"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	req, _ := http.NewRequest("GET", srv.URL+"/api/v1/materials", nil)
	_, err := c.Do(req)

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 503 {
		t.Errorf("expected 503, got %d", apiErr.StatusCode)
	}
	if apiErr.RequestID != "req-123" {
		t.Errorf("expected RequestID req-123, got %q", apiErr.RequestID)
	}
}

func TestWalkPages_AllItems(t *testing.T) {
	makeFixture := func(page, totalPages int) []byte {
		items := []map[string]interface{}{
			{"id": (page-1)*2 + 1, "name": "Material A"},
			{"id": (page-1)*2 + 2, "name": "Material B"},
		}
		data, _ := json.Marshal(map[string]interface{}{
			"materials": items,
			"meta": map[string]interface{}{
				"total_pages": totalPages,
				"total_count": totalPages * 2,
				"per_page":    2,
				"page":        page,
			},
		})
		return data
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		page := 1
		if pageStr != "" {
			fmt.Sscanf(pageStr, "%d", &page)
		}
		w.WriteHeader(200)
		w.Write(makeFixture(page, 3))
	}))
	defer srv.Close()

	c := newTestClient(srv)

	var emitted []json.RawMessage
	err := api.WalkPages(
		context.Background(),
		func(ctx context.Context, page int) ([]json.RawMessage, api.PageMeta, error) {
			reqURL := fmt.Sprintf("%s/api/v1/materials?page=%d", srv.URL, page)
			req, _ := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
			resp, err := c.Do(req)
			if err != nil {
				return nil, api.PageMeta{}, err
			}
			defer resp.Body.Close()

			var envelope struct {
				Materials []json.RawMessage `json:"materials"`
				Meta      api.RawPageMeta   `json:"meta"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
				return nil, api.PageMeta{}, err
			}
			meta := api.PageMeta{
				TotalPages: envelope.Meta.TotalPages,
				TotalCount: envelope.Meta.TotalCount,
				PerPage:    envelope.Meta.PerPage,
				Page:       envelope.Meta.CurrentPage,
			}
			return envelope.Materials, meta, nil
		},
		func(item json.RawMessage) {
			emitted = append(emitted, item)
		},
	)

	if err != nil {
		t.Fatalf("WalkPages failed: %v", err)
	}
	if len(emitted) != 6 {
		t.Errorf("expected 6 items, got %d", len(emitted))
	}
}

func TestWalkPages_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := api.WalkPages(ctx,
		func(ctx context.Context, page int) ([]json.RawMessage, api.PageMeta, error) {
			return nil, api.PageMeta{TotalPages: 5}, nil
		},
		func(item json.RawMessage) {},
	)

	if err == nil {
		t.Error("expected error from cancelled context, got nil")
	}
}
