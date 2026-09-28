package audit

import (
	"fmt"
	"os/exec"
)

// runMustangPDF invokes Mustang CLI when jar is configured (requires Java on PATH).
func runMustangPDF(pdfPath, jarPath string) ([]string, error) {
	java, err := exec.LookPath("java")
	if err != nil {
		return nil, fmt.Errorf("java not found for mustang validation: %w", err)
	}
	cmd := exec.Command(java, "-jar", jarPath, "--action", "validate", "--source", pdfPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("mustang validate failed: %w: %s", err, string(out))
	}
	notes := []string{"mustang validation completed"}
	if len(out) > 0 {
		notes = append(notes, string(out))
	}
	return notes, nil
}
