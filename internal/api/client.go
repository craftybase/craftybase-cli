package api

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/craftybase/stocksmith-cli/internal/brand"
)

type Client struct {
	BaseURL    string
	Token      string
	Version    string
	HTTPClient *http.Client
	Verbose    bool
}

func NewClient(baseURL, token, version string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		Version: version,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) userAgent() string {
	v := c.Version
	if v == "" {
		v = "dev"
	}
	return fmt.Sprintf("%s/%s (%s; %s)", brand.UserAgentProduct, v, runtime.GOOS, runtime.GOARCH)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.doWithRetry(req, 3)
}

func (c *Client) doWithRetry(req *http.Request, maxAttempts int) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("read request body: %w", err)
		}
		req.Body.Close()
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		c.setHeaders(req)

		if c.Verbose {
			fmt.Fprintf(os.Stderr, "> %s %s\n", req.Method, req.URL.String())
			fmt.Fprintf(os.Stderr, "> Authorization: Bearer ***REDACTED***\n")
			fmt.Fprintf(os.Stderr, "> User-Agent: %s\n", req.Header.Get("User-Agent"))
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			if attempt < maxAttempts-1 {
				wait := backoffDuration(attempt, 500*time.Millisecond)
				if c.Verbose {
					fmt.Fprintf(os.Stderr, "Network error: %v. Retrying in %s...\n", err, wait)
				}
				time.Sleep(wait)
				continue
			}
			return nil, fmt.Errorf("network error: %w. Run with --verbose for details", err)
		}

		if c.Verbose {
			fmt.Fprintf(os.Stderr, "< %s\n", resp.Status)
			if rid := resp.Header.Get("X-Request-Id"); rid != "" {
				fmt.Fprintf(os.Stderr, "< X-Request-Id: %s\n", rid)
			}
		}

		if resp.StatusCode == 429 {
			resp.Body.Close()
			waitDur := parse429Wait(resp, attempt)
			fmt.Fprintf(os.Stderr, "Rate limited. Waiting %s before retry...\n", waitDur.Round(time.Second))
			if attempt < maxAttempts-1 {
				time.Sleep(waitDur)
				continue
			}
			return nil, &APIError{StatusCode: 429, Message: "Rate limit exceeded. Try again later."}
		}

		if resp.StatusCode >= 400 {
			return nil, MapHTTPError(resp)
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded")
}

func (c *Client) DoRaw(req *http.Request) (*http.Response, error) {
	c.setHeaders(req)
	return c.HTTPClient.Do(req)
}

func (c *Client) setHeaders(req *http.Request) {
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("User-Agent", c.userAgent())
	req.Header.Set("Accept", "application/json")
}

func parse429Wait(resp *http.Response, attempt int) time.Duration {
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if secs, err := strconv.Atoi(ra); err == nil && secs > 0 {
			return time.Duration(secs) * time.Second
		}
	}
	fmt.Fprintf(os.Stderr, "Rate limited. Retry-After header absent; using backoff schedule.\n")
	wait := backoffDuration(attempt, time.Second)
	if wait < 60*time.Second {
		wait = 60 * time.Second
	}
	return wait
}

func backoffDuration(attempt int, base time.Duration) time.Duration {
	d := base * (1 << uint(attempt))
	jitter := time.Duration(float64(d) * 0.2 * (rand.Float64()*2 - 1))
	return d + jitter
}
