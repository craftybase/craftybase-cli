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
	"strings"

	"github.com/spf13/cobra"

	"github.com/craftybase/craftybase-cli/internal/api"
	"github.com/craftybase/craftybase-cli/internal/output"
)

// Project is a product or component (both are Projects on the API, distinguished
// by output_type). Only fields the CLI renders or passes through are modelled.
type Project struct {
	ID                 int           `json:"id"`
	Name               string        `json:"name"`
	SKU                string        `json:"sku"`
	OutputType         string        `json:"output_type"`
	Category           *string       `json:"category"`
	DefaultVariationID *int          `json:"default_variation_id"`
	Variations         []Variation   `json:"variations"`
	StockOnHand        string        `json:"stock_on_hand"`
	CommittedStock     string        `json:"committed_stock"`
	AvailableStock     string        `json:"available_stock"`
	LowStockLimit      string        `json:"low_stock_limit"`
	State              string        `json:"state"`
	UnitPrice          *output.Money `json:"unit_price"`
	CreatedAt          string        `json:"created_at"`
	UpdatedAt          string        `json:"updated_at"`
}

// Variation is an active variant of a Project.
type Variation struct {
	ID             int                  `json:"id"`
	Name           string               `json:"name"`
	SKU            string               `json:"sku"`
	Default        bool                 `json:"default"`
	StockOnHand    string               `json:"stock_on_hand"`
	CommittedStock string               `json:"committed_stock"`
	AvailableStock string               `json:"available_stock"`
	LowStockLimit  string               `json:"low_stock_limit"`
	State          string               `json:"state"`
	Attributes     []VariationAttribute `json:"attributes"`
	UnitPrice      *output.Money        `json:"unit_price"`
	CreatedAt      string               `json:"created_at"`
	UpdatedAt      string               `json:"updated_at"`
}

// VariationAttribute is one label/value pair on a Variation (e.g. Size: Large).
type VariationAttribute struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func categoryName(c *string) string {
	if c == nil || *c == "" {
		return "—"
	}
	return *c
}

func joinAttributes(attrs []VariationAttribute) string {
	if len(attrs) == 0 {
		return "—"
	}
	parts := make([]string, len(attrs))
	for i, a := range attrs {
		parts[i] = a.Label + ": " + a.Value
	}
	return strings.Join(parts, ", ")
}

func projectsToTable(raw []json.RawMessage) ([]string, [][]string) {
	headers := []string{"ID", "NAME", "SKU", "CATEGORY", "VARIANTS", "ON HAND", "AVAILABLE", "UNIT PRICE"}
	rows := make([][]string, 0, len(raw))
	for _, r := range raw {
		var p Project
		if err := json.Unmarshal(r, &p); err != nil {
			continue
		}
		rows = append(rows, projectToRow(&p))
	}
	return headers, rows
}

func projectToRow(p *Project) []string {
	return []string{
		strconv.Itoa(p.ID),
		p.Name,
		dashIfEmpty(p.SKU),
		categoryName(p.Category),
		strconv.Itoa(len(p.Variations)),
		p.StockOnHand,
		p.AvailableStock,
		output.FormatMoney(p.UnitPrice),
	}
}

func variationsToTable(vs []Variation) ([]string, [][]string) {
	headers := []string{"ID", "SKU", "ATTRIBUTES", "ON HAND", "AVAILABLE", "UNIT PRICE", "DEFAULT"}
	rows := make([][]string, 0, len(vs))
	for i := range vs {
		v := &vs[i]
		def := ""
		if v.Default {
			def = "✓"
		}
		rows = append(rows, []string{
			strconv.Itoa(v.ID),
			dashIfEmpty(v.SKU),
			joinAttributes(v.Attributes),
			v.StockOnHand,
			v.AvailableStock,
			output.FormatMoney(v.UnitPrice),
			def,
		})
	}
	return headers, rows
}

// renderProjectShow prints a single-row detail table, then a VARIATIONS (n)
// heading and a sub-table of the active variations (or an empty-state line).
func renderProjectShow(w io.Writer, p *Project, useColor bool) {
	headers := []string{"ID", "NAME", "SKU", "CATEGORY", "ON HAND", "AVAILABLE", "UNIT PRICE", "STATE"}
	row := []string{
		strconv.Itoa(p.ID),
		p.Name,
		dashIfEmpty(p.SKU),
		categoryName(p.Category),
		p.StockOnHand,
		p.AvailableStock,
		output.FormatMoney(p.UnitPrice),
		p.State,
	}
	output.FormatTable(w, headers, [][]string{row}, useColor)

	fmt.Fprintf(w, "\nVARIATIONS (%d)\n", len(p.Variations))
	if len(p.Variations) == 0 {
		fmt.Fprintln(w, "No active variations.")
		return
	}
	vh, vr := variationsToTable(p.Variations)
	output.FormatTable(w, vh, vr, useColor)
}

// projectResource configures the shared list/show runners for one resource.
type projectResource struct {
	pathSegment string // URL segment, e.g. "products"
	collection  string // list envelope key, e.g. "products"
	singular    string // show envelope key, e.g. "product"
}

type projectListFlags struct {
	sku, name, category, state string
	page, perPage              int
	all                        bool
}

func newProjectListCmd(res projectResource, f *projectListFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List " + res.collection,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectList(res, f)
		},
	}
	cmd.Flags().StringVar(&f.sku, "sku", "", "Filter by SKU (exact match)")
	cmd.Flags().StringVar(&f.name, "name", "", "Filter by name (substring match)")
	cmd.Flags().StringVar(&f.category, "category", "", "Filter by category name")
	cmd.Flags().StringVar(&f.state, "state", "", "Filter by state: active, archived, all")
	cmd.Flags().IntVar(&f.page, "page", 0, "Page number (1-based)")
	cmd.Flags().IntVar(&f.perPage, "per-page", 0, "Items per page (server clamps to 100)")
	cmd.Flags().BoolVar(&f.all, "all", false, "Fetch all pages and render as a single table")
	return cmd
}

func newProjectShowCmd(res projectResource) *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show a single " + res.singular,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectShow(res, args[0])
		},
	}
}

func runProjectList(res projectResource, f *projectListFlags) error {
	client, _, err := requireAuth()
	if err != nil {
		return err
	}

	if f.all && flagNDJSON {
		return fmt.Errorf("--all and --ndjson are mutually exclusive")
	}
	if f.all && f.page > 0 {
		return fmt.Errorf("--all and --page are mutually exclusive")
	}

	buildParams := func(page, perPage int) string {
		params := url.Values{}
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
		perPage := f.perPage
		if f.all && perPage == 0 {
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
		ctx := context.Background()
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

	if flagJSON && !f.all {
		page := f.page
		if page == 0 {
			page = 1
		}
		reqURL := client.BaseURL + "/api/v1/" + res.pathSegment + buildParams(page, f.perPage)
		req, err := http.NewRequest("GET", reqURL, nil)
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

	if f.all {
		var allItems []json.RawMessage
		ctx := context.Background()
		err := api.WalkPages(ctx, fetchPage, func(item json.RawMessage) {
			allItems = append(allItems, item)
		})
		if err != nil {
			return err
		}
		headers, rows := projectsToTable(allItems)
		output.FormatTable(os.Stdout, headers, rows, output.ColorEnabled(flagNoColor))
		fmt.Fprintf(os.Stdout, "(%d total)\n", len(allItems))
		return nil
	}

	page := f.page
	if page == 0 {
		page = 1
	}
	items, meta, err := fetchPage(context.Background(), page)
	if err != nil {
		return err
	}
	headers, rows := projectsToTable(items)
	output.FormatTable(os.Stdout, headers, rows, output.ColorEnabled(flagNoColor))
	fmt.Fprintf(os.Stdout, "(%d of %d)\n", len(items), meta.TotalCount)
	return nil
}

func runProjectShow(res projectResource, id string) error {
	client, _, err := requireAuth()
	if err != nil {
		return err
	}
	req, err := http.NewRequest("GET", client.BaseURL+"/api/v1/"+res.pathSegment+"/"+id, nil)
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
	var p Project
	if rawObj, ok := top[res.singular]; ok {
		if err := json.Unmarshal(rawObj, &p); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
	}
	renderProjectShow(os.Stdout, &p, output.ColorEnabled(flagNoColor))
	return nil
}
