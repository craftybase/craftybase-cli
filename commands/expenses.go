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

// Expense is a supplier purchase (the "purchases" data; there is no PurchaseOrder
// model). Only fields the CLI renders or passes through are modelled. Money fields
// are pointers — a JSON null decodes to nil → "—". Expenses gate the whole resource
// with a 403, so money is normally always present once past the gate.
type Expense struct {
	ID                 int               `json:"id"`
	Code               string            `json:"code"`
	PurchasedAt        string            `json:"purchased_at"`
	EstimatedArrivalAt string            `json:"estimated_arrival_at"`
	SupplierID         *int              `json:"supplier_id"`
	SupplierName       string            `json:"supplier_name"`
	Paid               bool              `json:"paid"`
	Received           bool              `json:"received"`
	Notes              string            `json:"notes"`
	Amount             *output.Money     `json:"amount"`
	ItemTotal          *output.Money     `json:"item_total"`
	Tax                *output.Money     `json:"tax"`
	Shipping           *output.Money     `json:"shipping"`
	Discount           *output.Money     `json:"discount"`
	LineItems          []ExpenseLineItem `json:"line_items"`
}

// ExpenseLineItem is one purchase line on an Expense. material_id is null for
// non-material (overhead/service/fee) lines.
type ExpenseLineItem struct {
	ID           int           `json:"id"`
	MaterialID   *int          `json:"material_id"`
	MaterialName string        `json:"material_name"`
	CategoryID   *int          `json:"category_id"`
	CategoryName string        `json:"category_name"`
	Quantity     string        `json:"quantity"`
	UnitPrice    *output.Money `json:"unit_price"`
	TotalPrice   *output.Money `json:"total_price"`
}

var expensesCmd = &cobra.Command{
	Use:   "expenses",
	Short: "Manage expenses",
}

var (
	expensesFilters    expenseFilters
	expensesPagination paginationFlags
)

// expenseFilters are the A6 list filters: from, to, updated_since, category_id,
// supplier_id. Values pass through verbatim; the API validates (HTTP 400).
type expenseFilters struct {
	from, to, updatedSince, categoryID, supplierID string
}

func (f *expenseFilters) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.from, "from", "", "Filter by purchase date on or after (ISO 8601, e.g. 2026-01-01)")
	cmd.Flags().StringVar(&f.to, "to", "", "Filter by purchase date on or before (ISO 8601)")
	cmd.Flags().StringVar(&f.updatedSince, "updated-since", "", "Return expenses updated on or after this time (ISO 8601; includes line-item edits)")
	cmd.Flags().StringVar(&f.categoryID, "category-id", "", "Filter by line-item category ID")
	cmd.Flags().StringVar(&f.supplierID, "supplier-id", "", "Filter by supplier ID")
}

func (f *expenseFilters) apply(params url.Values) {
	if f.from != "" {
		params.Set("from", f.from)
	}
	if f.to != "" {
		params.Set("to", f.to)
	}
	if f.updatedSince != "" {
		params.Set("updated_since", f.updatedSince)
	}
	if f.categoryID != "" {
		params.Set("category_id", f.categoryID)
	}
	if f.supplierID != "" {
		params.Set("supplier_id", f.supplierID)
	}
}

// expenseDate trims an ISO 8601 date or timestamp to YYYY-MM-DD; empty/short → "—".
// (Mirrors manufactureDate; consolidation is a deferred tidy.)
func expenseDate(ts string) string {
	if len(ts) < 10 {
		return "—"
	}
	return ts[:10]
}

func expensesToTable(rawItems []json.RawMessage) ([]string, [][]string) {
	headers := []string{"ID", "CODE", "SUPPLIER", "PURCHASED", "PAID", "RECEIVED", "ITEMS", "AMOUNT"}
	rows := make([][]string, 0, len(rawItems))
	for i, raw := range rawItems {
		var e Expense
		if err := json.Unmarshal(raw, &e); err != nil {
			warnSkip(i, err)
			continue
		}
		rows = append(rows, expenseToRow(&e))
	}
	return headers, rows
}

func expenseToRow(e *Expense) []string {
	return []string{
		strconv.Itoa(e.ID),
		dashIfEmpty(e.Code),
		refOrDash(e.SupplierName, e.SupplierID),
		expenseDate(e.PurchasedAt),
		output.FormatBool(e.Paid),
		output.FormatBool(e.Received),
		strconv.Itoa(len(e.LineItems)),
		output.FormatMoney(e.Amount),
	}
}

// renderExpenseShow prints a vertical key-value detail block, then a
// LINE ITEMS (n) heading and a sub-table of purchase lines (or empty state).
func renderExpenseShow(w io.Writer, raw json.RawMessage, useColor bool) error {
	var e Expense
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &e); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
	}

	output.FormatKeyValue(w, [][2]string{
		{"ID", strconv.Itoa(e.ID)},
		{"CODE", dashIfEmpty(e.Code)},
		{"SUPPLIER", refOrDash(e.SupplierName, e.SupplierID)},
		{"PURCHASED", expenseDate(e.PurchasedAt)},
		{"ESTIMATED ARRIVAL", expenseDate(e.EstimatedArrivalAt)},
		{"PAID", output.FormatBool(e.Paid)},
		{"RECEIVED", output.FormatBool(e.Received)},
		{"NOTES", dashIfEmpty(e.Notes)},
		{"AMOUNT", output.FormatMoney(e.Amount)},
		{"ITEM TOTAL", output.FormatMoney(e.ItemTotal)},
		{"TAX", output.FormatMoney(e.Tax)},
		{"SHIPPING", output.FormatMoney(e.Shipping)},
		{"DISCOUNT", output.FormatMoney(e.Discount)},
	}, useColor)

	fmt.Fprintf(w, "\nLINE ITEMS (%d)\n", len(e.LineItems))
	if len(e.LineItems) == 0 {
		fmt.Fprintln(w, "No line items.")
		return nil
	}
	headers := []string{"MATERIAL", "CATEGORY", "QTY", "UNIT PRICE", "TOTAL"}
	rows := make([][]string, 0, len(e.LineItems))
	for i := range e.LineItems {
		li := &e.LineItems[i]
		rows = append(rows, []string{
			refOrDash(li.MaterialName, li.MaterialID),
			refOrDash(li.CategoryName, li.CategoryID),
			dashIfEmpty(li.Quantity),
			output.FormatMoney(li.UnitPrice),
			output.FormatMoney(li.TotalPrice),
		})
	}
	output.FormatTable(w, headers, rows, useColor)
	return nil
}

func init() {
	res := resourceConfig{
		pathSegment: "expenses",
		collection:  "expenses",
		singular:    "expense",
		listLong: "List expenses (purchases) from your " + brand.ProductName + " account.\n\n" +
			"An expense is a supplier purchase — header totals plus the materials and costs\n" +
			"on each line. Filter by purchase-date range, change time, category, or supplier.\n" +
			"Use --all to fetch all pages, or --ndjson for streaming NDJSON output suitable\n" +
			"for data pipelines.",
		toTable:    expensesToTable,
		renderShow: renderExpenseShow,
	}
	expensesCmd.AddCommand(newResourceListCmd(res, &expensesFilters, &expensesPagination))
	expensesCmd.AddCommand(newResourceShowCmd(res))
	rootCmd.AddCommand(expensesCmd)
}
