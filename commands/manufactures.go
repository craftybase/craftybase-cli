package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/craftybase/stocksmith-cli/internal/brand"
	"github.com/craftybase/stocksmith-cli/internal/output"
)

// Manufacture is a production run. Only fields the CLI renders or passes through
// are modelled. Financial fields are pointers — absent (gated) decodes to nil.
type Manufacture struct {
	ID                 int                   `json:"id"`
	ProductID          int                   `json:"product_id"`
	VariationID        *int                  `json:"variation_id"`
	BatchCode          string                `json:"batch_code"`
	ProductionStatus   string                `json:"production_status"`
	Quantity           string                `json:"quantity"`
	ActualQuantity     string                `json:"actual_quantity"`
	MinutesWorked      int                   `json:"minutes_worked"`
	Notes              string                `json:"notes"`
	StartDate          string                `json:"start_date"`
	CompletedAt        string                `json:"completed_at"`
	DeadlineAt         string                `json:"deadline_at"`
	ExpiryAt           string                `json:"expiry_at"`
	LineItems          []ManufactureLineItem `json:"line_items"`
	TotalMaterialsCost *output.Money         `json:"total_materials_cost"`
	TotalLabourCost    *output.Money         `json:"total_labour_cost"`
}

// ManufactureLineItem is one consumed material on a Manufacture.
type ManufactureLineItem struct {
	ID          int           `json:"id"`
	MaterialID  int           `json:"material_id"`
	Name        string        `json:"name"`
	SKU         string        `json:"sku"`
	UnitMeasure string        `json:"unit_measure"`
	Quantity    string        `json:"quantity"`
	LotNumber   string        `json:"lot_number"`
	UnitPrice   *output.Money `json:"unit_price"`
}

var manufacturesCmd = &cobra.Command{
	Use:   "manufactures",
	Short: "Manage manufactures",
}

var (
	manufacturesFilters    manufactureFilters
	manufacturesPagination paginationFlags
)

// manufactureFilters are the A5 list filters: product_id, status, from, to.
type manufactureFilters struct {
	productID, status, from, to string
}

func (f *manufactureFilters) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.productID, "product-id", "", "Filter by product ID")
	cmd.Flags().StringVar(&f.status, "status", "", "Filter by production status: not_started, work_in_progress, completed")
	cmd.Flags().StringVar(&f.from, "from", "", "Filter by start date on or after (ISO 8601, e.g. 2026-01-01)")
	cmd.Flags().StringVar(&f.to, "to", "", "Filter by start date on or before (ISO 8601)")
}

func (f *manufactureFilters) apply(params url.Values) {
	if f.productID != "" {
		params.Set("product_id", f.productID)
	}
	if f.status != "" {
		params.Set("status", f.status)
	}
	if f.from != "" {
		params.Set("from", f.from)
	}
	if f.to != "" {
		params.Set("to", f.to)
	}
}

// manufactureDate trims an ISO 8601 timestamp to YYYY-MM-DD; empty/short → "—".
func manufactureDate(ts string) string {
	if len(ts) < 10 {
		return "—"
	}
	return ts[:10]
}

func manufacturesToTable(rawItems []json.RawMessage) ([]string, [][]string) {
	headers := []string{"ID", "PRODUCT", "STATUS", "QTY", "START", "ITEMS", "MATERIALS COST", "LABOUR COST"}
	rows := make([][]string, 0, len(rawItems))
	for i, raw := range rawItems {
		var m Manufacture
		if err := json.Unmarshal(raw, &m); err != nil {
			warnSkip(i, err)
			continue
		}
		rows = append(rows, manufactureToRow(&m))
	}
	return headers, rows
}

func manufactureToRow(m *Manufacture) []string {
	return []string{
		strconv.Itoa(m.ID),
		strconv.Itoa(m.ProductID),
		dashIfEmpty(m.ProductionStatus),
		dashIfEmpty(m.Quantity),
		manufactureDate(m.StartDate),
		strconv.Itoa(len(m.LineItems)),
		output.FormatMoney(m.TotalMaterialsCost),
		output.FormatMoney(m.TotalLabourCost),
	}
}

// renderManufactureShow prints a vertical key-value detail block, then a
// LINE ITEMS (n) heading and a sub-table of consumed materials (or empty state).
func renderManufactureShow(w io.Writer, raw json.RawMessage, useColor bool) error {
	var m Manufacture
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &m); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
	}

	variation := "—"
	if m.VariationID != nil {
		variation = strconv.Itoa(*m.VariationID)
	}
	output.FormatKeyValue(w, [][2]string{
		{"ID", strconv.Itoa(m.ID)},
		{"PRODUCT", strconv.Itoa(m.ProductID)},
		{"VARIATION", variation},
		{"BATCH", dashIfEmpty(m.BatchCode)},
		{"STATUS", dashIfEmpty(m.ProductionStatus)},
		{"PLANNED QTY", dashIfEmpty(m.Quantity)},
		{"ACTUAL QTY", dashIfEmpty(m.ActualQuantity)},
		{"MINUTES WORKED", strconv.Itoa(m.MinutesWorked)},
		{"START", manufactureDate(m.StartDate)},
		{"COMPLETED", manufactureDate(m.CompletedAt)},
		{"DEADLINE", manufactureDate(m.DeadlineAt)},
		{"EXPIRY", manufactureDate(m.ExpiryAt)},
		{"NOTES", dashIfEmpty(m.Notes)},
		{"MATERIALS COST", output.FormatMoney(m.TotalMaterialsCost)},
		{"LABOUR COST", output.FormatMoney(m.TotalLabourCost)},
	}, useColor)

	fmt.Fprintf(w, "\nLINE ITEMS (%d)\n", len(m.LineItems))
	if len(m.LineItems) == 0 {
		fmt.Fprintln(w, "No line items.")
		return nil
	}
	headers := []string{"MATERIAL", "NAME", "SKU", "QTY", "LOT", "UNIT PRICE"}
	rows := make([][]string, 0, len(m.LineItems))
	for i := range m.LineItems {
		li := &m.LineItems[i]
		qty := li.Quantity
		if li.UnitMeasure != "" {
			qty = li.Quantity + " " + li.UnitMeasure
		}
		rows = append(rows, []string{
			strconv.Itoa(li.MaterialID),
			li.Name,
			dashIfEmpty(li.SKU),
			qty,
			dashIfEmpty(li.LotNumber),
			output.FormatMoney(li.UnitPrice),
		})
	}
	output.FormatTable(w, headers, rows, useColor)
	return nil
}

func init() {
	res := resourceConfig{
		pathSegment: "manufactures",
		collection:  "manufactures",
		singular:    "manufacture",
		listLong: "List production runs (manufactures) from your " + brand.ProductName + " account.\n\n" +
			"Filter by product, status, or start-date range. Use --all to fetch all\n" +
			"pages, or --ndjson for streaming NDJSON output suitable for data pipelines.",
		toTable:    manufacturesToTable,
		renderShow: renderManufactureShow,
	}
	manufacturesCmd.AddCommand(newResourceListCmd(res, &manufacturesFilters, &manufacturesPagination))
	manufacturesCmd.AddCommand(newResourceShowCmd(res))
	rootCmd.AddCommand(manufacturesCmd)
}
