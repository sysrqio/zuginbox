package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		r.Messages = append(r.Messages, "java: not found on PATH — install a JRE 11+ for ZUGFeRD PDF validation (offline XML audit still works)")
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

// CheckWithPath is like Check but uses pathEnv only for java lookup (for tests).
func CheckWithPath(mustangJar, pathEnv string) Result {
	r := Result{JarPath: mustangJar}
	if p, ok := findJavaOnPath(pathEnv); ok {
		r.JavaOK = true
		r.JavaPath = p
	} else {
		r.Messages = append(r.Messages, "java: not found on PATH — install a JRE 11+ for ZUGFeRD PDF validation (offline XML audit still works)")
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

// ExitCode returns 0 when Java is available (and configured JAR exists when set), else 1.
func (r Result) ExitCode() int {
	if !r.JavaOK {
		return 1
	}
	if r.JarPath != "" && !r.JarOK {
		return 1
	}
	return 0
}

func findJavaOnPath(pathEnv string) (string, bool) {
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			continue
		}
		candidate := filepath.Join(dir, "java")
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			if st.Mode()&0o111 != 0 || strings.HasSuffix(candidate, ".exe") {
				return candidate, true
			}
		}
	}
	return "", false
}
