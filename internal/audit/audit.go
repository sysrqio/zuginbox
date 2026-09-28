package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sysrqio/zuginbox/internal/validator"
)

// Options configures a batch audit run.
type Options struct {
	Input         string
	MustangJar    string
	OutputJSON    string
	ReportMD      string
	FailOnWarn    bool
	Format        string // json | table
	OfflineFixture bool
}

// Receipt is one audited file entry in the report.
type Receipt struct {
	Path       string   `json:"path"`
	Kind       string   `json:"kind"` // xml | pdf
	Validation string   `json:"validation"`
	Issues     []string `json:"issues,omitempty"`
	Notes      []string `json:"notes,omitempty"`
	Fields     map[string]string `json:"fields,omitempty"`
}

// Report is the full audit output document.
type Report struct {
	GeneratedAt string    `json:"generated_at"`
	Tool        string    `json:"tool"`
	Mode        string    `json:"mode"`
	Receipts    []Receipt `json:"receipts"`
	Summary     Summary   `json:"summary"`
}

type Summary struct {
	Total   int `json:"total"`
	Valid   int `json:"valid"`
	Invalid int `json:"invalid"`
	Warn    int `json:"warn"`
}

// Run executes the audit and returns exit code: 0 ok, 1 IO/java, 2 invalid/warn with fail-on-warn.
func Run(opts Options) (int, error) {
	input := opts.Input
	if input == "" {
		return 1, fmt.Errorf("--input is required")
	}

	info, err := os.Stat(input)
	if err != nil {
		return 1, err
	}

	useMustang := opts.MustangJar != ""
	if useMustang {
		if _, err := os.Stat(opts.MustangJar); err != nil {
			return 1, fmt.Errorf("mustang jar: %w", err)
		}
	}

	mode := "offline-xml"
	if useMustang {
		mode = "mustang+offline"
	}
	if opts.OfflineFixture {
		mode = "offline-fixture"
	}

	var files []string
	if info.IsDir() {
		files, err = walkInvoiceFiles(input)
		if err != nil {
			return 1, err
		}
	} else {
		files = []string{input}
	}

	report := Report{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Tool:        "zuginbox",
		Mode:        mode,
	}

	for _, f := range files {
		rec, ioErr := auditFile(f, useMustang, opts.MustangJar)
		if ioErr != nil {
			return 1, ioErr
		}
		report.Receipts = append(report.Receipts, rec)
		switch rec.Validation {
		case "valid":
			report.Summary.Valid++
		case "invalid":
			report.Summary.Invalid++
		case "warn":
			report.Summary.Warn++
		}
		report.Summary.Total++
	}

	if opts.OutputJSON != "" {
		if err := writeJSON(opts.OutputJSON, report); err != nil {
			return 1, err
		}
	}
	if opts.ReportMD != "" {
		if err := writeMarkdown(opts.ReportMD, report); err != nil {
			return 1, err
		}
	}

	switch strings.ToLower(opts.Format) {
	case "table", "":
		printTable(report)
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(report)
	default:
		return 1, fmt.Errorf("unknown format %q", opts.Format)
	}

	if report.Summary.Invalid > 0 {
		return 2, nil
	}
	if opts.FailOnWarn && report.Summary.Warn > 0 {
		return 2, nil
	}
	return 0, nil
}

func walkInvoiceFiles(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".xml" || ext == ".pdf" {
			out = append(out, path)
		}
		return nil
	})
	return out, err
}

func auditFile(path string, useMustang bool, mustangJar string) (Receipt, error) {
	ext := strings.ToLower(filepath.Ext(path))
	rec := Receipt{
		Path:   path,
		Fields: make(map[string]string),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return rec, err
	}

	switch ext {
	case ".xml":
		rec.Kind = "xml"
		vr := validator.ValidateXML(data)
		rec.Fields = vr.Fields
		rec.Issues = append(rec.Issues, vr.Issues...)
		if vr.Valid {
			rec.Validation = "valid"
		} else {
			rec.Validation = "invalid"
		}
	case ".pdf":
		rec.Kind = "pdf"
		if useMustang {
			notes, javaErr := runMustangPDF(path, mustangJar)
			if javaErr != nil {
				return rec, javaErr
			}
			rec.Notes = append(rec.Notes, notes...)
			rec.Validation = "valid"
			if len(notes) > 0 {
				rec.Validation = "warn"
			}
		} else {
			rec.Validation = "warn"
			rec.Notes = append(rec.Notes,
				"PDF embedded XML not extracted in offline mode; install Java and Mustang JAR (--mustang-jar) for full ZUGFeRD validation",
			)
		}
	default:
		rec.Validation = "warn"
		rec.Notes = append(rec.Notes, "unsupported extension")
	}
	return rec, nil
}

func writeJSON(path string, report Report) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func writeMarkdown(path string, report Report) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# zuginbox audit report\n\n")
	fmt.Fprintf(&b, "- Generated: %s\n", report.GeneratedAt)
	fmt.Fprintf(&b, "- Mode: %s\n\n", report.Mode)
	fmt.Fprintf(&b, "## Summary\n\n")
	fmt.Fprintf(&b, "| Total | Valid | Invalid | Warn |\n")
	fmt.Fprintf(&b, "|------:|------:|--------:|-----:|\n")
	fmt.Fprintf(&b, "| %d | %d | %d | %d |\n\n", report.Summary.Total, report.Summary.Valid, report.Summary.Invalid, report.Summary.Warn)
	fmt.Fprintf(&b, "## Receipts\n\n")
	for _, r := range report.Receipts {
		fmt.Fprintf(&b, "### %s\n\n", r.Path)
		fmt.Fprintf(&b, "- Kind: %s\n", r.Kind)
		fmt.Fprintf(&b, "- Validation: **%s**\n", r.Validation)
		for _, iss := range r.Issues {
			fmt.Fprintf(&b, "- Issue: %s\n", iss)
		}
		for _, n := range r.Notes {
			fmt.Fprintf(&b, "- Note: %s\n", n)
		}
		b.WriteString("\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func printTable(report Report) {
	fmt.Printf("zuginbox audit (%s) — %d file(s): valid=%d invalid=%d warn=%d\n",
		report.Mode, report.Summary.Total, report.Summary.Valid, report.Summary.Invalid, report.Summary.Warn)
	for _, r := range report.Receipts {
		fmt.Printf("  %s [%s] %s\n", r.Path, r.Kind, r.Validation)
		for _, iss := range r.Issues {
			fmt.Printf("    ! %s\n", iss)
		}
		for _, n := range r.Notes {
			fmt.Printf("    ~ %s\n", n)
		}
	}
}
