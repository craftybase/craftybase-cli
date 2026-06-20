package commands_test

import (
	"encoding/json"
	"testing"
)

func componentListFixture() []byte {
	return []byte(`{
		"components": [
			{"id": 900, "name": "Wick Tab", "sku": "WICK-TAB",
			 "output_type": "component", "category": "Wicks",
			 "default_variation_id": 9001,
			 "stock_on_hand": "5000", "committed_stock": "200", "available_stock": "4800",
			 "low_stock_limit": "500", "state": "active",
			 "unit_price": {"amount": "0.05", "currency_code": "USD"},
			 "variations": []}
		],
		"meta": {"current_page": 1, "total_pages": 1, "total_count": 1, "per_page": 25}
	}`)
}

// Contract: components envelope uses "components", never "products"/"projects".
func TestComponentListFixture_ContractCheck(t *testing.T) {
	var raw map[string]interface{}
	if err := json.Unmarshal(componentListFixture(), &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["components"]; !ok {
		t.Fatal("envelope must use key \"components\"")
	}
	if _, ok := raw["products"]; ok {
		t.Error("components envelope must not contain a \"products\" key")
	}
	if _, ok := raw["projects"]; ok {
		t.Error("internal \"projects\" key must never leak")
	}
	comp := raw["components"].([]interface{})[0].(map[string]interface{})
	if comp["output_type"] != "component" {
		t.Errorf("output_type should be \"component\", got %v", comp["output_type"])
	}
}
