package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func fixtureRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join("..", "..", "testdata", "fixtures")
	if _, err := os.Stat(root); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestRun_validOnly(t *testing.T) {
	validDir := filepath.Join(fixtureRoot(t), "valid")
	out := filepath.Join(t.TempDir(), "out.json")
	code, err := Run(Options{
		Input:      validDir,
		OutputJSON: out,
		Format:     "json",
	})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Fatalf("exit code %d", code)
	}
}

func TestRun_mixedInvalid(t *testing.T) {
	root := fixtureRoot(t)
	out := filepath.Join(t.TempDir(), "out.json")
	code, err := Run(Options{
		Input:      root,
		OutputJSON: out,
		Format:     "table",
	})
	if err != nil {
		t.Fatal(err)
	}
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
}

func TestRun_failOnWarnPDF(t *testing.T) {
	// minimal fake pdf bytes
	dir := t.TempDir()
	pdf := filepath.Join(dir, "sample.pdf")
	if err := os.WriteFile(pdf, []byte("%PDF-1.4 minimal"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, err := Run(Options{
		Input:      pdf,
		FailOnWarn: true,
		OutputJSON: filepath.Join(dir, "out.json"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if code != 2 {
		t.Fatalf("expected exit 2 on warn, got %d", code)
	}
}

func TestWalkInvoiceFiles(t *testing.T) {
	files, err := walkInvoiceFiles(fixtureRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 3 {
		t.Fatalf("expected at least 3 files, got %d", len(files))
	}
}
