package commands

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func sampleProductJSON() []byte {
	return []byte(`{
		"id": 456, "name": "Beeswax Candle", "sku": "CDL-001",
		"output_type": "product", "category": "Candles",
		"default_variation_id": 4561,
		"stock_on_hand": "120", "committed_stock": "25", "available_stock": "95",
		"low_stock_limit": "10", "state": "active",
		"unit_price": {"amount": "24.00", "currency_code": "USD"},
		"variations": [
			{"id": 4561, "sku": "CDL-001-S", "name": "Small", "default": true,
			 "stock_on_hand": "40", "committed_stock": "10", "available_stock": "30",
			 "low_stock_limit": "5", "state": "active",
			 "attributes": [{"label": "Size", "value": "Small"}],
			 "unit_price": {"amount": "18.00", "currency_code": "USD"}},
			{"id": 4562, "sku": "CDL-001-M", "name": "Medium", "default": false,
			 "stock_on_hand": "50", "committed_stock": "5", "available_stock": "45",
			 "low_stock_limit": "5", "state": "active",
			 "attributes": [],
			 "unit_price": {"amount": "24.00", "currency_code": "USD"}}
		]
	}`)
}

func TestProjectToRow_Columns(t *testing.T) {
	var p Project
	if err := json.Unmarshal(sampleProductJSON(), &p); err != nil {
		t.Fatal(err)
	}
	row := projectToRow(&p)
	// ID, NAME, SKU, CATEGORY, VARIANTS, ON HAND, AVAILABLE, UNIT PRICE
	want := []string{"456", "Beeswax Candle", "CDL-001", "Candles", "2", "120", "95", "$24.00"}
	if len(row) != len(want) {
		t.Fatalf("expected %d columns, got %d: %v", len(want), len(row), row)
	}
	for i := range want {
		if row[i] != want[i] {
			t.Errorf("column %d: want %q, got %q", i, want[i], row[i])
		}
	}
}

func TestProjectToRow_NilPriceAndEmptySKURenderDash(t *testing.T) {
	p := Project{ID: 1, Name: "No Price", StockOnHand: "0", AvailableStock: "0"}
	row := projectToRow(&p)
	if row[2] != "—" {
		t.Errorf("empty SKU should render —, got %q", row[2])
	}
	if row[3] != "—" {
		t.Errorf("nil category should render —, got %q", row[3])
	}
	if row[7] != "—" {
		t.Errorf("nil unit_price should render —, got %q", row[7])
	}
}

func TestJoinAttributes(t *testing.T) {
	got := joinAttributes([]VariationAttribute{{Label: "Size", Value: "Large"}, {Label: "Scent", Value: "Vanilla"}})
	if got != "Size: Large, Scent: Vanilla" {
		t.Errorf("got %q", got)
	}
	if joinAttributes(nil) != "—" {
		t.Errorf("empty attributes should render —, got %q", joinAttributes(nil))
	}
}

func TestVariationsToTable_DefaultMarkerAndDash(t *testing.T) {
	var p Project
	if err := json.Unmarshal(sampleProductJSON(), &p); err != nil {
		t.Fatal(err)
	}
	headers, rows := variationsToTable(p.Variations)
	want := []string{"ID", "SKU", "ATTRIBUTES", "ON HAND", "AVAILABLE", "UNIT PRICE", "DEFAULT"}
	for i := range want {
		if headers[i] != want[i] {
			t.Errorf("header %d: want %q, got %q", i, want[i], headers[i])
		}
	}
	if rows[0][6] != "✓" {
		t.Errorf("default variation should show ✓, got %q", rows[0][6])
	}
	if rows[1][6] != "" {
		t.Errorf("non-default variation should be blank, got %q", rows[1][6])
	}
	if rows[1][2] != "—" {
		t.Errorf("variation with no attributes should render —, got %q", rows[1][2])
	}
}

func TestRenderProjectShow_DetailRowThenVariations(t *testing.T) {
	var p Project
	if err := json.Unmarshal(sampleProductJSON(), &p); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	renderProjectShow(&buf, &p, false)
	out := buf.String()

	for _, want := range []string{"Beeswax Candle", "STATE", "active", "VARIATIONS (2)", "CDL-001-S", "Size: Small", "✓"} {
		if !strings.Contains(out, want) {
			t.Errorf("show output missing %q\n---\n%s", want, out)
		}
	}
}

func TestRenderProjectShow_NoVariations(t *testing.T) {
	p := Project{ID: 7, Name: "Empty", State: "active", StockOnHand: "0", AvailableStock: "0"}
	var buf bytes.Buffer
	renderProjectShow(&buf, &p, false)
	out := buf.String()
	if !strings.Contains(out, "VARIATIONS (0)") {
		t.Errorf("expected VARIATIONS (0), got:\n%s", out)
	}
	if !strings.Contains(out, "No active variations.") {
		t.Errorf("expected empty-variations notice, got:\n%s", out)
	}
}
