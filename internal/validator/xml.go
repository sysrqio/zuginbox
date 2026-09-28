package validator

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Result holds lightweight XRechnung/ZUGFeRD-ish validation outcome.
type Result struct {
	Valid    bool
	Severity string // valid, invalid, warn
	Issues   []string
	Fields   map[string]string
}

// ValidateXML performs offline checks for common German e-invoice XML (CII/UBL fragments).
func ValidateXML(data []byte) Result {
	res := Result{
		Valid:    true,
		Severity: "valid",
		Fields:   make(map[string]string),
	}

	text := string(data)
	if len(strings.TrimSpace(text)) == 0 {
		return fail(res, "empty document")
	}

	if !xmlLooksWellFormed(data) {
		return fail(res, "malformed XML (not well-formed)")
	}

	// BT-1 Invoice number (must not pick guideline / context IDs)
	bt1 := findInvoiceNumber(text)
	if strings.TrimSpace(bt1) == "" {
		res.Valid = false
		res.Issues = append(res.Issues, "BT-1: missing invoice number (ID)")
	} else {
		res.Fields["BT-1"] = strings.TrimSpace(bt1)
	}

	// BT-24 Specification identifier (XRechnung guideline)
	bt24 := findFirst(text,
		`GuidelineSpecifiedDocumentContextParameter[\s\S]*?<ID>([^<]+)</ID>`,
		`<(?:[\w-]+:)?GuidelineSpecifiedDocumentContextParameter[\s\S]*?<(?:[\w-]+:)?ID>([^<]+)</`,
		`<ram:GuidelineSpecifiedDocumentContextParameter[\s\S]*?<ram:ID>([^<]+)</ram:ID>`,
	)
	if bt24 == "" {
		// UBL customization ID sometimes used
		bt24 = findFirst(text, `<(?:[\w-]+:)?CustomizationID>([^<]+)</`)
	}
	if strings.TrimSpace(bt24) == "" {
		res.Valid = false
		res.Issues = append(res.Issues, "BT-24: missing specification / guideline ID")
	} else {
		res.Fields["BT-24"] = strings.TrimSpace(bt24)
	}

	// Document currency total (BT-112-ish)
	total := findDocumentCurrencyTotal(text)
	if total == "" {
		res.Valid = false
		res.Issues = append(res.Issues, "missing document total (GrandTotalAmount / TaxInclusiveAmount / PayableAmount)")
	} else {
		res.Fields["DocumentCurrencyTotal"] = total
	}

	if !res.Valid {
		res.Severity = "invalid"
	}
	return res
}

func fail(res Result, msg string) Result {
	res.Valid = false
	res.Severity = "invalid"
	res.Issues = append(res.Issues, msg)
	return res
}

func xmlLooksWellFormed(data []byte) bool {
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			return true
		}
		if err != nil {
			return false
		}
	}
}

var rePatterns = make(map[string]*regexp.Regexp)

func re(pat string) *regexp.Regexp {
	if r, ok := rePatterns[pat]; ok {
		return r
	}
	r := regexp.MustCompile(pat)
	rePatterns[pat] = r
	return r
}

func findFirst(text string, patterns ...string) string {
	for _, p := range patterns {
		m := re(p).FindStringSubmatch(text)
		if len(m) >= 2 {
			return m[1]
		}
	}
	return ""
}

func findInvoiceNumber(text string) string {
	// Must not match ExchangedDocumentContext (prefix collision).
	if m := re(`<(?:[\w-]+:)?ExchangedDocument>([\s\S]*?)</(?:[\w-]+:)?ExchangedDocument>`).FindStringSubmatch(text); len(m) >= 2 {
		block := m[1]
		if id := findFirst(block, `<(?:[\w-]+:)?ID>([^<]+)</(?:[\w-]+:)?ID>`); id != "" {
			return id
		}
	}
	if id := findUBLInvoiceID(text); id != "" {
		return id
	}
	return ""
}

func findUBLInvoiceID(text string) string {
	// Invoice/cbc:ID near root Invoice element
	if m := re(`<(?:[\w-]+:)?Invoice[\s\S]*?<(?:[\w-]+:)?ID>([^<]+)</(?:[\w-]+:)?ID>`).FindStringSubmatch(text); len(m) >= 2 {
		return m[1]
	}
	return ""
}

func findDocumentCurrencyTotal(text string) string {
	patterns := []string{
		`<(?:[\w-]+:)?GrandTotalAmount[^>]*>([^<]+)</`,
		`<(?:[\w-]+:)?TaxInclusiveAmount[^>]*>([^<]+)</`,
		`<(?:[\w-]+:)?PayableAmount[^>]*>([^<]+)</`,
		`<(?:[\w-]+:)?LineTotalAmount[^>]*>([^<]+)</`,
	}
	for _, p := range patterns {
		if v := findFirst(text, p); v != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// CompareTotals returns an issue string if a and b differ (normalized decimal compare).
func CompareTotals(a, b string) string {
	na := normalizeAmount(a)
	nb := normalizeAmount(b)
	if na == "" || nb == "" {
		return ""
	}
	if na != nb {
		return fmt.Sprintf("total mismatch: XML=%s vs other=%s", na, nb)
	}
	return ""
}

func normalizeAmount(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.TrimSpace(s)
	return s
}
