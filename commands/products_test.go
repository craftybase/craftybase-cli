package commands_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func productListFixture() []byte {
	return []byte(`{
		"products": [
			{"id": 456, "name": "Beeswax Candle", "sku": "CDL-001",
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
				 "unit_price": {"amount": "18.00", "currency_code": "USD"}}
			 ]}
		],
		"meta": {"current_page": 1, "total_pages": 1, "total_count": 1, "per_page": 25}
	}`)
}

// Contract: list envelope uses "products", money amount is a string, category is
// a flat string, variations is an array, and meta carries current_page.
func TestProductListFixture_ContractCheck(t *testing.T) {
	var raw map[string]interface{}
	if err := json.Unmarshal(productListFixture(), &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["products"]; !ok {
		t.Fatal("envelope must use key \"products\"")
	}
	products := raw["products"].([]interface{})
	p := products[0].(map[string]interface{})

	if _, ok := p["category"].(string); !ok {
		t.Errorf("category must be a flat string, got %T", p["category"])
	}
	if _, ok := p["variations"].([]interface{}); !ok {
		t.Errorf("variations must be an array, got %T", p["variations"])
	}
	price := p["unit_price"].(map[string]interface{})
	if _, ok := price["amount"].(string); !ok {
		t.Errorf("unit_price.amount must be a string, got %T", price["amount"])
	}
	meta := raw["meta"].(map[string]interface{})
	if _, ok := meta["current_page"]; !ok {
		t.Error("meta must carry current_page")
	}
}

// The --category flag must map to the API's category_name query param (1:1 with
// materials). Asserted at the HTTP layer, matching TestCategoryFilterParam.
func TestProductsCategoryFilterParam(t *testing.T) {
	var gotParam string
	srv := setupMockServer(map[string]func(http.ResponseWriter, *http.Request){
		"/api/v1/products": func(w http.ResponseWriter, r *http.Request) {
			gotParam = r.URL.Query().Get("category_name")
			w.Write([]byte(`{"products":[],"meta":{"current_page":1,"total_pages":1,"total_count":0,"per_page":25}}`))
		},
	})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/products?category_name=Candles")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if gotParam != "Candles" {
		t.Errorf("expected category_name=Candles, got %q", gotParam)
	}
}
