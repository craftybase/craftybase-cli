package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/craftybase/stocksmith-cli/internal/api"
	"github.com/craftybase/stocksmith-cli/internal/output"
)

// warnWriter receives non-fatal warnings (e.g. skipped malformed list items).
// It is os.Stderr in production; tests may redirect it.
var warnWriter io.Writer = os.Stderr

// warnSkip reports a list item that could not be decoded, so a dropped row is
// never silent.
func warnSkip(index int, err error) {
	fmt.Fprintf(warnWriter, "warning: skipping malformed item at index %d: %v\n", index, err)
}

// resourceConfig configures the shared list/show runners for one resource.
type resourceConfig struct {
	pathSegment string // URL segment under /api/v1/, e.g. "materials"
	collection  string // list envelope key, e.g. "materials"
	singular    string // show envelope key, e.g. "material"
	listLong    string // optional Long help for `list` (empty => Cobra uses Short)

	toTable    func(raw []json.RawMessage) (headers []string, rows [][]string)
	renderShow func(w io.Writer, raw json.RawMessage, useColor bool) error
}

// listFilters supplies a resource's filter flags and their query contribution.
// The runner owns pagination (page/per_page/all); a listFilters owns everything
// resource-specific (e.g. --sku, --status).
type listFilters interface {
	addFlags(cmd *cobra.Command)
	apply(params url.Values)
}

// paginationFlags are the universal list flags every resource shares.
type paginationFlags struct {
	page, perPage int
	all           bool
}

// projectFilters are the shared filters for the project-family resources
// (materials, products, components).
type projectFilters struct {
	sku, name, category, state string
}

func (f *projectFilters) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.sku, "sku", "", "Filter by SKU (exact match)")
	cmd.Flags().StringVar(&f.name, "name", "", "Filter by name (substring match)")
	cmd.Flags().StringVar(&f.category, "category", "", "Filter by category name")
	cmd.Flags().StringVar(&f.state, "state", "", "Filter by state: active, archived, all")
}

func (f *projectFilters) apply(params url.Values) {
	if f.sku != "" {
		params.Set("sku", f.sku)
	}
	if f.name != "" {
		params.Set("name", f.name)
	}
	if f.category != "" {
		params.Set("category_name", f.category)
	}
	if f.state != "" {
		params.Set("state", f.state)
	}
}

func newResourceListCmd(res resourceConfig, filters listFilters, pag *paginationFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List " + res.collection,
		Long:  res.listLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResourceList(cmd.Context(), res, filters, pag)
		},
	}
	filters.addFlags(cmd)
	cmd.Flags().IntVar(&pag.page, "page", 0, "Page number (1-based)")
	cmd.Flags().IntVar(&pag.perPage, "per-page", 0, "Items per page (server clamps to 100)")
	cmd.Flags().BoolVar(&pag.all, "all", false, "Fetch all pages and render as a single table")
	return cmd
}

func newResourceShowCmd(res resourceConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show a single " + res.singular,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResourceShow(cmd.Context(), res, args[0])
		},
	}
}

// validateListFlags rejects the output-mode / pagination combinations the list
// runner cannot honor. Pure (no auth or HTTP) so the precedence is unit-testable.
func validateListFlags(jsonOut, ndjson, all bool, page int) error {
	if all && ndjson {
		return fmt.Errorf("--all and --ndjson are mutually exclusive")
	}
	if all && page > 0 {
		return fmt.Errorf("--all and --page are mutually exclusive")
	}
	if jsonOut && all {
		return fmt.Errorf("--json and --all are mutually exclusive (use --ndjson to stream all pages as JSON)")
	}
	return nil
}

func runResourceList(ctx context.Context, res resourceConfig, filters listFilters, pag *paginationFlags) error {
	if err := validateListFlags(flagJSON, flagNDJSON, pag.all, pag.page); err != nil {
		return err
	}

	client, _, err := requireAuth()
	if err != nil {
		return err
	}

	buildParams := func(page, perPage int) string {
		params := url.Values{}
		filters.apply(params)
		if page > 0 {
			params.Set("page", strconv.Itoa(page))
		}
		if perPage > 0 {
			params.Set("per_page", strconv.Itoa(perPage))
		}
		q := params.Encode()
		if q != "" {
			return "?" + q
		}
		return ""
	}

	fetchPage := func(ctx context.Context, page int) ([]json.RawMessage, api.PageMeta, error) {
		perPage := pag.perPage
		if pag.all && perPage == 0 {
			perPage = 100
		}
		reqURL := client.BaseURL + "/api/v1/" + res.pathSegment + buildParams(page, perPage)
		req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return nil, api.PageMeta{}, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, api.PageMeta{}, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, api.PageMeta{}, err
		}

		top := map[string]json.RawMessage{}
		if err := json.Unmarshal(body, &top); err != nil {
			return nil, api.PageMeta{}, fmt.Errorf("parse response: %w", err)
		}
		var items []json.RawMessage
		if rawList, ok := top[res.collection]; ok {
			if err := json.Unmarshal(rawList, &items); err != nil {
				return nil, api.PageMeta{}, fmt.Errorf("parse response: %w", err)
			}
		}
		var rawMeta api.RawPageMeta
		if rawM, ok := top["meta"]; ok {
			_ = json.Unmarshal(rawM, &rawMeta)
		}
		meta := api.PageMeta{
			TotalPages: rawMeta.TotalPages,
			TotalCount: rawMeta.TotalCount,
			PerPage:    rawMeta.PerPage,
			Page:       rawMeta.CurrentPage,
		}
		return items, meta, nil
	}

	if flagNDJSON {
		var total int
		err := api.WalkPages(ctx, fetchPage, func(item json.RawMessage) {
			total++
			output.WriteNDJSONLine(os.Stdout, item) //nolint:errcheck
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "(%d total)\n", total)
		return nil
	}

	if flagJSON {
		page := pag.page
		if page == 0 {
			page = 1
		}
		reqURL := client.BaseURL + "/api/v1/" + res.pathSegment + buildParams(page, pag.perPage)
		req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return output.PrintJSON(os.Stdout, body)
	}

	if pag.all {
		var allItems []json.RawMessage
		err := api.WalkPages(ctx, fetchPage, func(item json.RawMessage) {
			allItems = append(allItems, item)
		})
		if err != nil {
			return err
		}
		headers, rows := res.toTable(allItems)
		output.FormatTable(os.Stdout, headers, rows, output.ColorEnabled(flagNoColor))
		fmt.Fprintf(os.Stdout, "(%d total)\n", len(allItems))
		return nil
	}

	page := pag.page
	if page == 0 {
		page = 1
	}
	items, meta, err := fetchPage(ctx, page)
	if err != nil {
		return err
	}
	headers, rows := res.toTable(items)
	output.FormatTable(os.Stdout, headers, rows, output.ColorEnabled(flagNoColor))
	fmt.Fprintf(os.Stdout, "(%d of %d)\n", len(items), meta.TotalCount)
	return nil
}

func runResourceShow(ctx context.Context, res resourceConfig, id string) error {
	client, _, err := requireAuth()
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", client.BaseURL+"/api/v1/"+res.pathSegment+"/"+id, nil)
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

	top := map[string]json.RawMessage{}
	if err := json.Unmarshal(body, &top); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	return res.renderShow(os.Stdout, top[res.singular], output.ColorEnabled(flagNoColor))
}
