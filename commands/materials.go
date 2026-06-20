package commands

import (
	"encoding/json"
	"io"
	"strconv"

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

var materialsListFlags resourceListFlags

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

// renderMaterialShow adapts the one-row materials table to the
// resourceConfig.renderShow signature.
func renderMaterialShow(w io.Writer, raw json.RawMessage, useColor bool) error {
	headers, rows := materialsToTable([]json.RawMessage{raw})
	output.FormatTable(w, headers, rows, useColor)
	return nil
}

func init() {
	res := resourceConfig{
		pathSegment: "materials",
		collection:  "materials",
		singular:    "material",
		listLong: `List materials from your Craftybase account.

Filter by SKU, name, category, or state. Use --all to fetch all pages,
or --ndjson for streaming NDJSON output suitable for data pipelines.`,
		toTable:    materialsToTable,
		renderShow: renderMaterialShow,
	}
	materialsCmd.AddCommand(newResourceListCmd(res, &materialsListFlags))
	materialsCmd.AddCommand(newResourceShowCmd(res))
	rootCmd.AddCommand(materialsCmd)
}
