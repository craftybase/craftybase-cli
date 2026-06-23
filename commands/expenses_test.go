package commands

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

func sampleExpenseJSON() json.RawMessage {
	return json.RawMessage(`{
		"id": 88, "code": "EXP-001A",
		"purchased_at": "2026-05-14T09:00:00Z", "estimated_arrival_at": "2026-05-21",
		"supplier_id": 12, "supplier_name": "Candle Supply Co.",
		"paid": true, "received": false, "notes": "Bulk restock",
		"amount":     {"amount": "312.50", "currency_code": "USD"},
		"item_total": {"amount": "290.00", "currency_code": "USD"},
		"tax":        {"amount": "10.00",  "currency_code": "USD"},
		"shipping":   {"amount": "12.50",  "currency_code": "USD"},
		"discount":   {"amount": "0.00",   "currency_code": "USD"},
		"line_items": [
			{"id": 401, "material_id": 5, "material_name": "Soy Wax", "material_expense": true,
			 "category_id": 3, "category_name": "Raw Materials", "quantity": "50.0",
			 "unit_price": {"amount": "4.00", "currency_code": "USD"},
			 "total_price": {"amount": "200.00", "currency_code": "USD"}},
			{"id": 403, "material_id": null, "material_name": null, "material_expense": false,
			 "category_id": 7, "category_name": "Freight & Handling", "quantity": "1.0",
			 "unit_price": {"amount": "12.50", "currency_code": "USD"},
			 "total_price": {"amount": "12.50", "currency_code": "USD"}}
		]
	}`)
}

func TestExpenseDate(t *testing.T) {
	if got := expenseDate("2026-05-14T09:00:00Z"); got != "2026-05-14" {
		t.Errorf("timestamp trim: want 2026-05-14, got %q", got)
	}
	if got := expenseDate("2026-05-21"); got != "2026-05-21" {
		t.Errorf("date-only: want 2026-05-21, got %q", got)
	}
	if got := expenseDate(""); got != "—" {
		t.Errorf("empty should render —, got %q", got)
	}
}

func TestExpenseFilters_Apply(t *testing.T) {
	f := expenseFilters{from: "2026-01-01", to: "2026-06-30", updatedSince: "2026-06-01T00:00:00Z", categoryID: "3", supplierID: "12"}
	params := url.Values{}
	f.apply(params)
	want := map[string]string{"from": "2026-01-01", "to": "2026-06-30", "updated_since": "2026-06-01T00:00:00Z", "category_id": "3", "supplier_id": "12"}
	for k, v := range want {
		if params.Get(k) != v {
			t.Errorf("param %q: want %q, got %q", k, v, params.Get(k))
		}
	}
	if len(params) != len(want) {
		t.Errorf("expected exactly %d params, got %d: %v", len(want), len(params), params)
	}
	empty := url.Values{}
	(&expenseFilters{}).apply(empty)
	if len(empty) != 0 {
		t.Errorf("empty filters should set no params, got %v", empty)
	}
}

func TestExpensesToTable_Columns(t *testing.T) {
	headers, rows := expensesToTable([]json.RawMessage{sampleExpenseJSON()})
	wantHeaders := []string{"ID", "CODE", "SUPPLIER", "PURCHASED", "PAID", "RECEIVED", "ITEMS", "AMOUNT"}
	for i, h := range wantHeaders {
		if headers[i] != h {
			t.Errorf("header %d: want %q, got %q", i, h, headers[i])
		}
	}
	want := []string{"88", "EXP-001A", "Candle Supply Co.", "2026-05-14", "yes", "no", "2", "$312.50"}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	for i := range want {
		if rows[0][i] != want[i] {
			t.Errorf("col %d: want %q, got %q", i, want[i], rows[0][i])
		}
	}
}

func TestExpensesToTable_FallbacksAndSkip(t *testing.T) {
	// null code + null supplier (name+id) → "—"; absent amount → "—"; malformed sibling skipped.
	var buf bytes.Buffer
	orig := warnWriter
	warnWriter = &buf
	defer func() { warnWriter = orig }()
	raw := json.RawMessage(`{"id":9,"code":null,"supplier_id":null,"supplier_name":null,"paid":false,"received":true,"purchased_at":"2026-02-02T00:00:00Z","line_items":[]}`)
	_, rows := expensesToTable([]json.RawMessage{raw, json.RawMessage(`{bad`)})
	if len(rows) != 1 {
		t.Fatalf("want 1 row (malformed skipped), got %d", len(rows))
	}
	want := []string{"9", "—", "—", "2026-02-02", "no", "yes", "0", "—"}
	for i := range want {
		if rows[0][i] != want[i] {
			t.Errorf("col %d: want %q, got %q", i, want[i], rows[0][i])
		}
	}
	if !strings.Contains(buf.String(), "skipping malformed item") {
		t.Errorf("expected warning, got %q", buf.String())
	}
}

func TestRenderExpenseShow_DetailAndLineItems(t *testing.T) {
	var buf bytes.Buffer
	if err := renderExpenseShow(&buf, sampleExpenseJSON(), false); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"ID", "88", "CODE", "EXP-001A", "SUPPLIER", "Candle Supply Co.",
		"PURCHASED", "2026-05-14", "ESTIMATED ARRIVAL", "2026-05-21",
		"PAID", "yes", "RECEIVED", "no", "NOTES", "Bulk restock",
		"AMOUNT", "$312.50", "ITEM TOTAL", "$290.00", "TAX", "$10.00",
		"SHIPPING", "$12.50", "DISCOUNT", "$0.00",
		"LINE ITEMS (2)", "MATERIAL", "CATEGORY", "QTY", "UNIT PRICE", "TOTAL",
		"Soy Wax", "Raw Materials", "50.0", "$4.00", "$200.00",
		"Freight & Handling",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("show output missing %q\n---\n%s", want, out)
		}
	}
}

func TestRenderExpenseShow_NoLineItems(t *testing.T) {
	raw := json.RawMessage(`{"id":7,"code":"EXP-Z","paid":false,"received":false,"purchased_at":"2026-01-01T00:00:00Z","line_items":[]}`)
	var buf bytes.Buffer
	if err := renderExpenseShow(&buf, raw, false); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "LINE ITEMS (0)") || !strings.Contains(out, "No line items.") {
		t.Errorf("expected empty line-items state, got:\n%s", out)
	}
}
