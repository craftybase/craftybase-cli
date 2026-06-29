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

// Recipe is a bill of materials (formulation) for a product or variation. Only
// fields the CLI renders or passes through are modelled; financial fields are
// pointers — absent (gated for view-only keys) or JSON null both decode to nil.
type Recipe struct {
	ID                       int                `json:"id"`
	ProductID                int                `json:"product_id"`
	ProductName              string             `json:"product_name"`
	VariationID              *int               `json:"variation_id"`
	VariationName            string             `json:"variation_name"`
	Name                     string             `json:"name"`
	ManufactureBatchQuantity string             `json:"manufacture_batch_quantity"`
	ManufactureMinutes       int                `json:"manufacture_minutes"`
	Notes                    string             `json:"notes"`
	Ingredients              []RecipeIngredient `json:"ingredients"`
	TotalCost                *output.Money      `json:"total_cost"`
	UnitCost                 *output.Money      `json:"unit_cost"`
	TotalLabourCost          *output.Money      `json:"total_labour_cost"`
	UnitLabourCost           *output.Money      `json:"unit_labour_cost"`
	TotalCOGS                *output.Money      `json:"total_cogs"`
	UnitCOGS                 *output.Money      `json:"unit_cogs"`
}

// RecipeIngredient is one consumed material (bill-of-materials line) on a Recipe.
// Ingredients carry no per-line cost — costs exist only as recipe-level rollups.
type RecipeIngredient struct {
	ID           int    `json:"id"`
	MaterialID   int    `json:"material_id"`
	MaterialName string `json:"material_name"`
	Quantity     string `json:"quantity"`
	Unit         string `json:"unit"`
}

var recipesCmd = &cobra.Command{
	Use:   "recipes",
	Short: "Manage recipes",
}

var (
	recipesFilters    recipeFilters
	recipesPagination paginationFlags
)

// recipeFilters are the A4 list filters: product_id, variation_id, updated_since.
// Values pass through verbatim; the API validates (HTTP 400 on a malformed value).
type recipeFilters struct {
	productID, variationID, updatedSince string
}

func (f *recipeFilters) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.productID, "product-id", "", "Filter by product ID")
	cmd.Flags().StringVar(&f.variationID, "variation-id", "", "Filter by variation ID")
	cmd.Flags().StringVar(&f.updatedSince, "updated-since", "", "Return recipes updated on or after this time (ISO 8601, e.g. 2026-01-01)")
}

func (f *recipeFilters) apply(params url.Values) {
	if f.productID != "" {
		params.Set("product_id", f.productID)
	}
	if f.variationID != "" {
		params.Set("variation_id", f.variationID)
	}
	if f.updatedSince != "" {
		params.Set("updated_since", f.updatedSince)
	}
}

// nameOrID returns name when non-empty, else the id as a string. Used to show a
// human label (product / material) and fall back to the id when the name is null.
func nameOrID(name string, id int) string {
	if name != "" {
		return name
	}
	return strconv.Itoa(id)
}

// refOrDash renders a nullable reference: name, else id, else "—". Shared by
// recipes (variation) and expenses (supplier, material, category).
func refOrDash(name string, id *int) string {
	if name != "" {
		return name
	}
	if id != nil {
		return strconv.Itoa(*id)
	}
	return "—"
}

func recipesToTable(rawItems []json.RawMessage) ([]string, [][]string) {
	headers := []string{"ID", "PRODUCT", "NAME", "VARIATION", "BATCH", "INGR", "UNIT COST", "UNIT COGS"}
	rows := make([][]string, 0, len(rawItems))
	for i, raw := range rawItems {
		var r Recipe
		if err := json.Unmarshal(raw, &r); err != nil {
			warnSkip(i, err)
			continue
		}
		rows = append(rows, recipeToRow(&r))
	}
	return headers, rows
}

func recipeToRow(r *Recipe) []string {
	return []string{
		strconv.Itoa(r.ID),
		nameOrID(r.ProductName, r.ProductID),
		dashIfEmpty(r.Name),
		refOrDash(r.VariationName, r.VariationID),
		dashIfEmpty(r.ManufactureBatchQuantity),
		strconv.Itoa(len(r.Ingredients)),
		output.FormatMoney(r.UnitCost),
		output.FormatMoney(r.UnitCOGS),
	}
}

// renderRecipeShow prints a vertical key-value detail block, then an
// INGREDIENTS (n) heading and a sub-table of bill-of-materials lines (or the
// empty state).
func renderRecipeShow(w io.Writer, raw json.RawMessage, useColor bool) error {
	var r Recipe
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &r); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
	}

	output.FormatKeyValue(w, [][2]string{
		{"ID", strconv.Itoa(r.ID)},
		{"PRODUCT", nameOrID(r.ProductName, r.ProductID)},
		{"VARIATION", refOrDash(r.VariationName, r.VariationID)},
		{"NAME", dashIfEmpty(r.Name)},
		{"BATCH QTY", dashIfEmpty(r.ManufactureBatchQuantity)},
		{"MINUTES", strconv.Itoa(r.ManufactureMinutes)},
		{"NOTES", dashIfEmpty(r.Notes)},
		{"TOTAL COST", output.FormatMoney(r.TotalCost)},
		{"UNIT COST", output.FormatMoney(r.UnitCost)},
		{"TOTAL LABOUR", output.FormatMoney(r.TotalLabourCost)},
		{"UNIT LABOUR", output.FormatMoney(r.UnitLabourCost)},
		{"TOTAL COGS", output.FormatMoney(r.TotalCOGS)},
		{"UNIT COGS", output.FormatMoney(r.UnitCOGS)},
	}, useColor)

	fmt.Fprintf(w, "\nINGREDIENTS (%d)\n", len(r.Ingredients))
	if len(r.Ingredients) == 0 {
		fmt.Fprintln(w, "No ingredients.")
		return nil
	}
	headers := []string{"MATERIAL", "QTY", "UNIT"}
	rows := make([][]string, 0, len(r.Ingredients))
	for i := range r.Ingredients {
		ing := &r.Ingredients[i]
		rows = append(rows, []string{
			nameOrID(ing.MaterialName, ing.MaterialID),
			dashIfEmpty(ing.Quantity),
			dashIfEmpty(ing.Unit),
		})
	}
	output.FormatTable(w, headers, rows, useColor)
	return nil
}

func init() {
	res := resourceConfig{
		pathSegment: "recipes",
		collection:  "recipes",
		singular:    "recipe",
		listLong: "List recipes (bills of materials) from your " + brand.ProductName + " account.\n\n" +
			"A recipe is the formulation for a product or variation — the materials and\n" +
			"quantities consumed per batch, with cost and COGS rollups. Filter by product,\n" +
			"variation, or change time. Use --all to fetch all pages, or --ndjson for\n" +
			"streaming NDJSON output suitable for data pipelines.",
		toTable:    recipesToTable,
		renderShow: renderRecipeShow,
	}
	recipesCmd.AddCommand(newResourceListCmd(res, &recipesFilters, &recipesPagination))
	recipesCmd.AddCommand(newResourceShowCmd(res))
	rootCmd.AddCommand(recipesCmd)
}
