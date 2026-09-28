package version

// Version is set at build time via -ldflags.
var Version = "dev"

// Commit is set at build time via -ldflags.
var Commit = "none"

// Date is set at build time via -ldflags.
var Date = "unknown"

func String() string {
	return Version
}

func Full() string {
	return Version + " (commit " + Commit + ", built " + Date + ")"
}
