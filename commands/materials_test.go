package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/craftybase/craftybase-cli/internal/output"
)

func sampleMaterialJSON() []byte {
	return []byte(`{
		"id": 123, "name": "Organic Beeswax", "sku": "WAX-001",
		"category": "Waxes", "stock_on_hand": "12.5", "unit_measure": "kg",
		"unit_cost": {"amount": "8.75", "currency_code": "USD"}
	}`)
}

func TestRenderMaterialShow_SingleRowTable(t *testing.T) {
	var buf bytes.Buffer
	if err := renderMaterialShow(&buf, sampleMaterialJSON(), false); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"ID", "NAME", "SKU", "CATEGORY", "ON HAND", "UNIT COST", "Organic Beeswax", "WAX-001", "Waxes", "12.5 kg", "$8.75"} {
		if !strings.Contains(out, want) {
			t.Errorf("show output missing %q\n---\n%s", want, out)
		}
	}
}

func TestMaterialToRow_NilCostEmptySKUNilCategoryRenderDash(t *testing.T) {
	m := Material{ID: 1, Name: "No Price", StockOnHand: "0", UnitCost: (*output.Money)(nil)}
	row := materialToRow(&m)
	// ID, NAME, SKU, CATEGORY, ON HAND, UNIT COST
	if row[2] != "—" {
		t.Errorf("empty SKU should render —, got %q", row[2])
	}
	if row[3] != "—" {
		t.Errorf("nil category should render —, got %q", row[3])
	}
	if row[5] != "—" {
		t.Errorf("nil unit_cost should render —, got %q", row[5])
	}
}
