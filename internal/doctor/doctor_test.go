package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExitCode_javaMissing(t *testing.T) {
	r := CheckWithPath("", "")
	if r.ExitCode() != 1 {
		t.Fatalf("expected exit 1 without java, got %d", r.ExitCode())
	}
}

func TestExitCode_javaOnPathOverride(t *testing.T) {
	dir := t.TempDir()
	java := filepath.Join(dir, "java")
	if err := os.WriteFile(java, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := CheckWithPath("", dir)
	if !r.JavaOK {
		t.Fatal("expected java ok with PATH override")
	}
	if r.ExitCode() != 0 {
		t.Fatalf("expected exit 0, got %d", r.ExitCode())
	}
}

func TestExitCode_missingJar(t *testing.T) {
	dir := t.TempDir()
	java := filepath.Join(dir, "java")
	if err := os.WriteFile(java, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	jar := filepath.Join(dir, "missing.jar")
	r := CheckWithPath(jar, dir)
	if r.ExitCode() != 1 {
		t.Fatalf("expected exit 1 for missing jar, got %d", r.ExitCode())
	}
}
