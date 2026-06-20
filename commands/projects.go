package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

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
	for i, r := range raw {
		var p Project
		if err := json.Unmarshal(r, &p); err != nil {
			warnSkip(i, err)
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

// renderProjectShowRaw adapts renderProjectShow to the resourceConfig.renderShow
// signature. An empty raw (missing envelope key) renders a zero-value Project,
// preserving the prior behavior.
func renderProjectShowRaw(w io.Writer, raw json.RawMessage, useColor bool) error {
	var p Project
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &p); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
	}
	renderProjectShow(w, &p, useColor)
	return nil
}
