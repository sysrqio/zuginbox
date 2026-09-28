package doctor

import (
	"fmt"
	"os"
	"os/exec"
)

// Result summarizes environment checks.
type Result struct {
	JavaOK    bool
	JavaPath  string
	JarOK     bool
	JarPath   string
	Messages  []string
}

// Check verifies Java and optional Mustang JAR presence.
func Check(mustangJar string) Result {
	r := Result{JarPath: mustangJar}
	if p, err := exec.LookPath("java"); err == nil {
		r.JavaOK = true
		r.JavaPath = p
	} else {
		r.Messages = append(r.Messages, "java: not found on PATH (optional for offline XML audit)")
	}

	if mustangJar == "" {
		r.Messages = append(r.Messages, "mustang jar: not configured (use --mustang-jar or scripts/fetch-mustang.sh)")
		return r
	}
	if _, err := os.Stat(mustangJar); err == nil {
		r.JarOK = true
	} else {
		r.Messages = append(r.Messages, fmt.Sprintf("mustang jar: missing at %s", mustangJar))
	}
	return r
}

func (r Result) Print() {
	fmt.Println("zuginbox doctor")
	if r.JavaOK {
		fmt.Printf("  java: ok (%s)\n", r.JavaPath)
	} else {
		fmt.Println("  java: missing")
	}
	if r.JarPath != "" {
		if r.JarOK {
			fmt.Printf("  mustang jar: ok (%s)\n", r.JarPath)
		} else {
			fmt.Printf("  mustang jar: missing (%s)\n", r.JarPath)
		}
	} else {
		fmt.Println("  mustang jar: not specified")
	}
	for _, m := range r.Messages {
		fmt.Printf("  note: %s\n", m)
	}
}
