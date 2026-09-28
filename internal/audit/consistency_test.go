package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyConsistencyChecks_mismatchFixture(t *testing.T) {
	dir := filepath.Join(fixtureRoot(t), "consistency-mismatch")
	files, err := walkInvoiceFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	var entries []receiptEntry
	for _, f := range files {
		rec, err := auditFile(f, false, "")
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, receiptEntry{path: f, rec: rec})
	}
	failures := applyConsistencyChecks(entries)
	if failures < 1 {
		t.Fatalf("expected consistency failures >= 1, got %d", failures)
	}
	var pdf *Receipt
	for i := range entries {
		if entries[i].rec.Kind == "pdf" {
			pdf = &entries[i].rec
			break
		}
	}
	if pdf == nil || pdf.Consistency == nil || pdf.Consistency.Status != "fail" {
		t.Fatalf("pdf consistency: %+v", pdf)
	}
}

func TestExtractPDFTotal(t *testing.T) {
	pdf := filepath.Join(fixtureRoot(t), "consistency-mismatch", "invoice.pdf")
	data, err := os.ReadFile(pdf)
	if err != nil {
		t.Fatal(err)
	}
	if got := extractPDFTotal(data); got != "120.00" {
		t.Fatalf("pdf total %q", got)
	}
}

func TestRun_consistencyMismatchExit2(t *testing.T) {
	dir := filepath.Join(fixtureRoot(t), "consistency-mismatch")
	out := filepath.Join(t.TempDir(), "out.json")
	code, err := Run(Options{
		Input:      dir,
		OutputJSON: out,
		Format:     "json",
	})
	if err != nil {
		t.Fatal(err)
	}
	if code != 2 {
		t.Fatalf("expected exit 2 for consistency mismatch, got %d", code)
	}
}
