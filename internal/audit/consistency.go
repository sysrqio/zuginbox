package audit

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sysrqio/zuginbox/internal/validator"
)

var pdfTotalPatterns = []*regexp.Regexp{
	regexp.MustCompile(`PDF-TOTAL:([0-9][0-9.,]*)`),
	regexp.MustCompile(`Total payable:\s*([0-9][0-9.,]*)`),
	regexp.MustCompile(`GrandTotal:\s*([0-9][0-9.,]*)`),
}

// Consistency holds PDF/XML total comparison for one receipt pair.
type Consistency struct {
	Status   string `json:"status"` // ok, fail, skip
	XMLTotal string `json:"xml_total,omitempty"`
	PDFTotal string `json:"pdf_total,omitempty"`
}

func applyConsistencyChecks(entries []receiptEntry) int {
	byStem := map[string]map[string]*receiptEntry{}
	for i := range entries {
		e := &entries[i]
		ext := strings.ToLower(filepath.Ext(e.path))
		if ext != ".xml" && ext != ".pdf" {
			continue
		}
		stem := strings.TrimSuffix(e.path, ext)
		if byStem[stem] == nil {
			byStem[stem] = map[string]*receiptEntry{}
		}
		byStem[stem][ext] = e
	}

	failures := 0
	for _, pair := range byStem {
		xmlEnt := pair[".xml"]
		pdfEnt := pair[".pdf"]
		if xmlEnt == nil || pdfEnt == nil {
			continue
		}
		xmlTotal := xmlEnt.rec.Fields["DocumentCurrencyTotal"]
		pdfData, err := os.ReadFile(pdfEnt.path)
		if err != nil {
			continue
		}
		pdfTotal := extractPDFTotal(pdfData)
		c := Consistency{Status: "skip", XMLTotal: xmlTotal, PDFTotal: pdfTotal}
		if xmlTotal == "" || pdfTotal == "" {
			c.Status = "skip"
		} else if issue := validator.CompareTotals(xmlTotal, pdfTotal); issue != "" {
			c.Status = "fail"
			failures++
			pdfEnt.rec.Issues = append(pdfEnt.rec.Issues, "consistency: "+issue)
			if pdfEnt.rec.Validation == "valid" {
				pdfEnt.rec.Validation = "warn"
			}
		} else {
			c.Status = "ok"
		}
		pdfEnt.rec.Consistency = &c
		xmlEnt.rec.Consistency = &c
	}
	return failures
}

func extractPDFTotal(data []byte) string {
	text := string(data)
	for _, p := range pdfTotalPatterns {
		if m := p.FindStringSubmatch(text); len(m) >= 2 {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}
