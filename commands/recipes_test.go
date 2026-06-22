package commands

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

func sampleRecipeJSON() json.RawMessage {
	return json.RawMessage(`{
		"id": 42, "product_id": 7, "product_name": "Soy Candle",
		"variation_id": null, "variation_name": null,
		"name": "Base Recipe", "manufacture_batch_quantity": "10.0",
		"manufacture_minutes": 60, "notes": "Pour at 60C",
		"ingredients": [
			{"id": 5001, "material_id": 312, "material_name": "Soy Wax", "quantity": "200.0", "unit": "grams"},
			{"id": 5002, "material_id": 313, "material_name": "Cotton Wick", "quantity": "1.0", "unit": "each"}
		],
		"total_cost": {"amount": "50.00", "currency_code": "USD"},
		"unit_cost": {"amount": "5.00", "currency_code": "USD"},
		"total_labour_cost": {"amount": "15.00", "currency_code": "USD"},
		"unit_labour_cost": {"amount": "1.50", "currency_code": "USD"},
		"total_cogs": {"amount": "65.00", "currency_code": "USD"},
		"unit_cogs": {"amount": "6.50", "currency_code": "USD"}
	}`)
}

func TestNameOrID(t *testing.T) {
	if got := nameOrID("Soy Wax", 312); got != "Soy Wax" {
		t.Errorf("want name, got %q", got)
	}
	if got := nameOrID("", 312); got != "312" {
		t.Errorf("empty name should fall back to id, got %q", got)
	}
}

func TestVariationRef(t *testing.T) {
	id := 3
	if got := variationRef("Small", &id); got != "Small" {
		t.Errorf("want name, got %q", got)
	}
	if got := variationRef("", &id); got != "3" {
		t.Errorf("empty name should fall back to id, got %q", got)
	}
	if got := variationRef("", nil); got != "—" {
		t.Errorf("nil id should render —, got %q", got)
	}
}

func TestRecipeFilters_Apply(t *testing.T) {
	f := recipeFilters{productID: "7", variationID: "3", updatedSince: "2026-01-01"}
	params := url.Values{}
	f.apply(params)
	want := map[string]string{"product_id": "7", "variation_id": "3", "updated_since": "2026-01-01"}
	for k, v := range want {
		if params.Get(k) != v {
			t.Errorf("param %q: want %q, got %q", k, v, params.Get(k))
		}
	}
	if len(params) != len(want) {
		t.Errorf("expected exactly %d params, got %d: %v", len(want), len(params), params)
	}
	empty := url.Values{}
	(&recipeFilters{}).apply(empty)
	if len(empty) != 0 {
		t.Errorf("empty filters should set no params, got %v", empty)
	}
}

func TestRecipesToTable_Columns(t *testing.T) {
	headers, rows := recipesToTable([]json.RawMessage{sampleRecipeJSON()})
	wantHeaders := []string{"ID", "PRODUCT", "NAME", "VARIATION", "BATCH", "INGR", "UNIT COST", "UNIT COGS"}
	for i, h := range wantHeaders {
		if headers[i] != h {
			t.Errorf("header %d: want %q, got %q", i, h, headers[i])
		}
	}
	want := []string{"42", "Soy Candle", "Base Recipe", "—", "10.0", "2", "$5.00", "$6.50"}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	for i := range want {
		if rows[0][i] != want[i] {
			t.Errorf("col %d: want %q, got %q", i, want[i], rows[0][i])
		}
	}
}

func TestRecipesToTable_GatedCostsAndFallbacks(t *testing.T) {
	// view-only key: cost fields absent; product_name null → id fallback; variation present.
	raw := json.RawMessage(`{"id":9,"product_id":7,"product_name":null,
		"variation_id":3,"variation_name":"Small","name":"V",
		"manufacture_batch_quantity":"1.0","ingredients":[]}`)
	_, rows := recipesToTable([]json.RawMessage{raw})
	want := []string{"9", "7", "V", "Small", "1.0", "0", "—", "—"}
	for i := range want {
		if rows[0][i] != want[i] {
			t.Errorf("col %d: want %q, got %q", i, want[i], rows[0][i])
		}
	}
}

func TestRecipesToTable_SkipsMalformed(t *testing.T) {
	var buf bytes.Buffer
	orig := warnWriter
	warnWriter = &buf
	defer func() { warnWriter = orig }()
	_, rows := recipesToTable([]json.RawMessage{json.RawMessage(`{bad`)})
	if len(rows) != 0 {
		t.Errorf("malformed item should be skipped, got %d rows", len(rows))
	}
	if !strings.Contains(buf.String(), "skipping malformed item") {
		t.Errorf("expected warning, got %q", buf.String())
	}
}

func TestRenderRecipeShow_DetailAndIngredients(t *testing.T) {
	var buf bytes.Buffer
	if err := renderRecipeShow(&buf, sampleRecipeJSON(), false); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"ID", "42", "PRODUCT", "Soy Candle", "VARIATION", "—", "NAME", "Base Recipe",
		"BATCH QTY", "10.0", "MINUTES", "60", "NOTES", "Pour at 60C",
		"TOTAL COST", "$50.00", "UNIT COST", "$5.00",
		"TOTAL LABOUR", "$15.00", "UNIT LABOUR", "$1.50",
		"TOTAL COGS", "$65.00", "UNIT COGS", "$6.50",
		"INGREDIENTS (2)", "MATERIAL", "QTY", "UNIT",
		"Soy Wax", "200.0", "grams", "Cotton Wick",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("show output missing %q\n---\n%s", want, out)
		}
	}
}

func TestRenderRecipeShow_NoIngredients(t *testing.T) {
	raw := json.RawMessage(`{"id":7,"product_id":3,"product_name":"X","name":"Empty","manufacture_batch_quantity":"0","ingredients":[]}`)
	var buf bytes.Buffer
	if err := renderRecipeShow(&buf, raw, false); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "INGREDIENTS (0)") || !strings.Contains(out, "No ingredients.") {
		t.Errorf("expected empty ingredients state, got:\n%s", out)
	}
}

func TestRenderRecipeShow_GatedCostsRenderDash(t *testing.T) {
	raw := json.RawMessage(`{"id":7,"product_id":3,"product_name":"X","name":"Y","manufacture_batch_quantity":"1.0","ingredients":[]}`)
	var buf bytes.Buffer
	if err := renderRecipeShow(&buf, raw, false); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, label := range []string{"TOTAL COST", "UNIT COST", "TOTAL LABOUR", "UNIT LABOUR", "TOTAL COGS", "UNIT COGS"} {
		if !strings.Contains(out, label) {
			t.Errorf("missing cost label %q", label)
		}
	}
	if !strings.Contains(out, "—") {
		t.Errorf("gated costs should render —, got:\n%s", out)
	}
}
