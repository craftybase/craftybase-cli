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

	"github.com/craftybase/craftybase-cli/internal/api"
	"github.com/craftybase/craftybase-cli/internal/output"
	"github.com/spf13/cobra"
)

type Material struct {
	ID          int           `json:"id"`
	Name        string        `json:"name"`
	SKU         string        `json:"sku"`
	Category    *string       `json:"category"`
	StockOnHand string        `json:"stock_on_hand"`
	UnitMeasure string        `json:"unit_measure"`
	UnitCost    *output.Money `json:"unit_cost"`
}

var materialsCmd = &cobra.Command{
	Use:   "materials",
	Short: "Manage materials",
}

var (
	matSKU     string
	matName    string
	matCat     string
	matState   string
	matPage    int
	matPerPage int
	flagAll    bool
)

var materialsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List materials",
	Long: `List materials from your Craftybase account.

Filter by SKU, name, category, or state. Use --all to fetch all pages,
or --ndjson for streaming NDJSON output suitable for data pipelines.`,
	RunE: runMaterialsList,
}

func runMaterialsList(cmd *cobra.Command, args []string) error {
	client, _, err := requireAuth()
	if err != nil {
		return err
	}

	if flagAll && flagNDJSON {
		return fmt.Errorf("--all and --ndjson are mutually exclusive")
	}
	if flagAll && matPage > 0 {
		return fmt.Errorf("--all and --page are mutually exclusive")
	}

	buildParams := func(page, perPage int) string {
		params := url.Values{}
		if matSKU != "" {
			params.Set("sku", matSKU)
		}
		if matName != "" {
			params.Set("name", matName)
		}
		if matCat != "" {
			params.Set("category_name", matCat)
		}
		if matState != "" {
			params.Set("state", matState)
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
		perPage := matPerPage
		if flagAll && perPage == 0 {
			perPage = 100
		}
		reqURL := client.BaseURL + "/api/v1/materials" + buildParams(page, perPage)
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

		var envelope struct {
			Materials []json.RawMessage `json:"materials"`
			Meta      api.RawPageMeta   `json:"meta"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			return nil, api.PageMeta{}, fmt.Errorf("parse response: %w", err)
		}

		meta := api.PageMeta{
			TotalPages: envelope.Meta.TotalPages,
			TotalCount: envelope.Meta.TotalCount,
			PerPage:    envelope.Meta.PerPage,
			Page:       envelope.Meta.CurrentPage,
		}
		return envelope.Materials, meta, nil
	}

	if flagNDJSON {
		ctx := context.Background()
		var totalEmitted int
		err := api.WalkPages(ctx, fetchPage, func(item json.RawMessage) {
			totalEmitted++
			output.WriteNDJSONLine(os.Stdout, item) //nolint:errcheck
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "(%d total)\n", totalEmitted)
		return nil
	}

	if flagJSON && !flagAll {
		page := matPage
		if page == 0 {
			page = 1
		}
		perPage := matPerPage
		reqURL := client.BaseURL + "/api/v1/materials" + buildParams(page, perPage)
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

	if flagAll {
		var allItems []json.RawMessage
		ctx := context.Background()
		err := api.WalkPages(ctx, fetchPage, func(item json.RawMessage) {
			allItems = append(allItems, item)
		})
		if err != nil {
			return err
		}

		headers, rows := materialsToTable(allItems)
		output.FormatTable(os.Stdout, headers, rows, output.ColorEnabled(flagNoColor))
		fmt.Fprintf(os.Stdout, "(%d total)\n", len(allItems))
		return nil
	}

	page := matPage
	if page == 0 {
		page = 1
	}
	items, meta, err := fetchPage(context.Background(), page)
	if err != nil {
		return err
	}

	headers, rows := materialsToTable(items)
	output.FormatTable(os.Stdout, headers, rows, output.ColorEnabled(flagNoColor))
	fmt.Fprintf(os.Stdout, "(%d of %d)\n", len(items), meta.TotalCount)
	return nil
}

func materialsToTable(rawItems []json.RawMessage) ([]string, [][]string) {
	headers := []string{"ID", "NAME", "SKU", "CATEGORY", "ON HAND", "UNIT COST"}
	rows := make([][]string, 0, len(rawItems))

	for _, raw := range rawItems {
		var m Material
		if err := json.Unmarshal(raw, &m); err != nil {
			continue
		}
		rows = append(rows, materialToRow(&m))
	}
	return headers, rows
}

func materialToRow(m *Material) []string {
	sku := m.SKU
	if sku == "" {
		sku = "—"
	}
	category := "—"
	if m.Category != nil && *m.Category != "" {
		category = *m.Category
	}
	onHand := m.StockOnHand
	if m.UnitMeasure != "" {
		onHand = m.StockOnHand + " " + m.UnitMeasure
	}
	unitCost := output.FormatMoney(m.UnitCost)

	return []string{
		strconv.Itoa(m.ID),
		m.Name,
		sku,
		category,
		onHand,
		unitCost,
	}
}

var materialsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a single material",
	Args:  cobra.ExactArgs(1),
	RunE:  runMaterialsShow,
}

func runMaterialsShow(cmd *cobra.Command, args []string) error {
	client, _, err := requireAuth()
	if err != nil {
		return err
	}

	id := args[0]
	req, err := http.NewRequest("GET", client.BaseURL+"/api/v1/materials/"+id, nil)
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

	var envelope struct {
		Material json.RawMessage `json:"material"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	headers, rows := materialsToTable([]json.RawMessage{envelope.Material})
	output.FormatTable(os.Stdout, headers, rows, output.ColorEnabled(flagNoColor))
	return nil
}

func init() {
	materialsListCmd.Flags().StringVar(&matSKU, "sku", "", "Filter by SKU (exact match)")
	materialsListCmd.Flags().StringVar(&matName, "name", "", "Filter by name (substring match)")
	materialsListCmd.Flags().StringVar(&matCat, "category", "", "Filter by category name")
	materialsListCmd.Flags().StringVar(&matState, "state", "", "Filter by state: active, archived, all")
	materialsListCmd.Flags().IntVar(&matPage, "page", 0, "Page number (1-based)")
	materialsListCmd.Flags().IntVar(&matPerPage, "per-page", 0, "Items per page (server clamps to 100)")
	materialsListCmd.Flags().BoolVar(&flagAll, "all", false, "Fetch all pages and render as a single table")

	materialsCmd.AddCommand(materialsListCmd)
	materialsCmd.AddCommand(materialsShowCmd)
	rootCmd.AddCommand(materialsCmd)
}
