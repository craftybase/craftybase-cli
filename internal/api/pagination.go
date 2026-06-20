package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type PageMeta struct {
	TotalPages int
	TotalCount int
	PerPage    int
	Page       int
}

type RawPageMeta struct {
	TotalPages  int `json:"total_pages"`
	TotalCount  int `json:"total_count"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
}

func WalkPages(
	ctx context.Context,
	fetchPage func(ctx context.Context, page int) (items []json.RawMessage, meta PageMeta, err error),
	emit func(item json.RawMessage),
) error {
	totalPages := 1
	for page := 1; page <= totalPages; page++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		pageCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		items, meta, err := fetchWithRetry(pageCtx, page, fetchPage)
		cancel()

		if err != nil {
			return fmt.Errorf("page %d: %w", page, err)
		}

		if page == 1 {
			totalPages = meta.TotalPages
			if totalPages < 1 {
				totalPages = 1
			}
		}

		for _, item := range items {
			emit(item)
		}
	}
	return nil
}

func fetchWithRetry(
	ctx context.Context,
	page int,
	fetchPage func(ctx context.Context, page int) ([]json.RawMessage, PageMeta, error),
) ([]json.RawMessage, PageMeta, error) {
	maxRetries := 5
	for attempt := 0; attempt < maxRetries; attempt++ {
		items, meta, err := fetchPage(ctx, page)
		if err == nil {
			return items, meta, nil
		}

		apiErr, ok := err.(*APIError)
		if !ok || apiErr.StatusCode != 429 {
			return nil, PageMeta{}, err
		}

		if attempt == maxRetries-1 {
			return nil, PageMeta{}, fmt.Errorf("rate limit exceeded after %d retries for page %d", maxRetries, page)
		}

		wait := backoffDuration(attempt, time.Second)
		if wait < 60*time.Second {
			wait = 60 * time.Second
		}
		fmt.Fprintf(os.Stderr, "Rate limited on page %d. Waiting %s...\n", page, wait.Round(time.Second))

		select {
		case <-ctx.Done():
			return nil, PageMeta{}, ctx.Err()
		case <-time.After(wait):
		}
	}
	return nil, PageMeta{}, fmt.Errorf("unexpected retry loop exit")
}
