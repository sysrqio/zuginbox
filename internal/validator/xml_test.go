package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateXML_validFixture(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "fixtures", "valid", "minimal-xrechnung-cii.xml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	r := ValidateXML(data)
	if !r.Valid {
		t.Fatalf("expected valid, got issues: %v", r.Issues)
	}
	if r.Fields["BT-1"] != "RE-2026-001" {
		t.Fatalf("BT-1: %q", r.Fields["BT-1"])
	}
	if r.Fields["DocumentCurrencyTotal"] != "119.00" {
		t.Fatalf("total: %q", r.Fields["DocumentCurrencyTotal"])
	}
}

func TestValidateXML_invalidMissingBT1(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "fixtures", "invalid", "missing-bt1.xml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	r := ValidateXML(data)
	if r.Valid {
		t.Fatal("expected invalid")
	}
}

func TestValidateXML_malformed(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "fixtures", "invalid", "malformed.xml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	r := ValidateXML(data)
	if r.Valid {
		t.Fatal("expected invalid")
	}
}

func TestCompareTotals(t *testing.T) {
	if CompareTotals("119.00", "119,00") != "" {
		t.Fatal("expected match")
	}
	if CompareTotals("119.00", "120.00") == "" {
		t.Fatal("expected mismatch")
	}
}
