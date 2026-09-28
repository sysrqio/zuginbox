package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func buildCLI(t *testing.T) string {
	t.Helper()
	root := repoRoot(t)
	out := filepath.Join(t.TempDir(), "zuginbox")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", out, ".")
	cmd.Dir = filepath.Join(root, "cmd", "zuginbox")
	if combined, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build zuginbox: %v\n%s", err, combined)
	}
	return out
}

type execOutput struct {
	b []byte
}

func (e *execOutput) Write(p []byte) (int, error) {
	e.b = append(e.b, p...)
	return len(p), nil
}

func (e *execOutput) String() string {
	return string(e.b)
}

func runCLI(t *testing.T, bin string, workDir string, env []string, args ...string) (stdout string, exitCode int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Env = env
	var outBuf execOutput
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("run zuginbox: %v", err)
		}
	}
	return outBuf.String(), exitCode
}

func TestCLIVersion(t *testing.T) {
	bin := buildCLI(t)
	out, code := runCLI(t, bin, repoRoot(t), os.Environ(), "version")
	if code != 0 {
		t.Fatalf("version exit %d", code)
	}
	if !strings.Contains(out, "dev") && !strings.Contains(out, "0.") {
		t.Fatalf("unexpected version output: %q", out)
	}
}

func TestCLIAuditOfflineFixture(t *testing.T) {
	bin := buildCLI(t)
	root := repoRoot(t)
	dir := t.TempDir()
	outJSON := filepath.Join(dir, "audit.json")
	out, code := runCLI(t, bin, root, os.Environ(),
		"audit", "--offline-fixture", "--format", "json", "--output", outJSON,
	)
	if code != 2 {
		t.Fatalf("offline fixtures include invalid samples; expected exit 2, got %d\n%s", code, out)
	}
	if _, err := os.Stat(outJSON); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"tool": "zuginbox"`) {
		t.Fatalf("missing json tool field:\n%s", out)
	}
}

func TestCLIDoctorJavaMissing(t *testing.T) {
	bin := buildCLI(t)
	_, code := runCLI(t, bin, repoRoot(t), []string{"PATH=/nonexistent"}, "doctor")
	if code != 1 {
		t.Fatalf("expected doctor exit 1 without java, got %d", code)
	}
}
