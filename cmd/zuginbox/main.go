package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sysrqio/zuginbox/internal/audit"
	"github.com/sysrqio/zuginbox/internal/doctor"
	"github.com/sysrqio/zuginbox/internal/version"
)

func main() {
	if err := newRoot().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "zuginbox",
		Short: "Local batch audit for German e-invoices (ZUGFeRD / XRechnung)",
	}

	root.AddCommand(newAuditCmd())
	root.AddCommand(newVersionCmd())
	root.AddCommand(newDoctorCmd())
	return root
}

func newAuditCmd() *cobra.Command {
	var (
		input          string
		mustangJar     string
		outputJSON     string
		reportMD       string
		failOnWarn     bool
		format         string
		offlineFixture bool
	)

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Audit XML/PDF invoices in a directory or single file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if offlineFixture {
				if input == "" {
					input = defaultFixtureDir()
				}
			}
			code, err := audit.Run(audit.Options{
				Input:          input,
				MustangJar:     mustangJar,
				OutputJSON:     outputJSON,
				ReportMD:       reportMD,
				FailOnWarn:     failOnWarn,
				Format:         format,
				OfflineFixture: offlineFixture,
			})
			if err != nil {
				return err
			}
			if code != 0 {
				os.Exit(code)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&input, "input", "", "Directory or file to audit (required unless --offline-fixture)")
	cmd.Flags().StringVar(&mustangJar, "mustang-jar", "", "Path to Mustang project JAR (optional)")
	cmd.Flags().StringVar(&outputJSON, "output", "receipt-audit.json", "Write JSON report to this path")
	cmd.Flags().StringVar(&reportMD, "report-md", "receipt-audit.md", "Write Markdown report to this path")
	cmd.Flags().BoolVar(&failOnWarn, "fail-on-warn", false, "Exit 2 when any receipt is warn")
	cmd.Flags().StringVar(&format, "format", "table", "Stdout format: json|table")
	cmd.Flags().BoolVar(&offlineFixture, "offline-fixture", false, "Audit bundled test fixtures (demo / CI)")
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if offlineFixture && !cmd.Flags().Changed("input") {
			input = defaultFixtureDir()
			return nil
		}
		if input == "" {
			return fmt.Errorf("--input is required")
		}
		return nil
	}
	return cmd
}

func defaultFixtureDir() string {
	for _, p := range []string{"testdata/fixtures", "../testdata/fixtures"} {
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	return "testdata/fixtures"
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Full())
		},
	}
}

func newDoctorCmd() *cobra.Command {
	var mustangJar string
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check Java and Mustang JAR availability",
		Run: func(cmd *cobra.Command, args []string) {
			doctor.Check(mustangJar).Print()
		},
	}
	cmd.Flags().StringVar(&mustangJar, "mustang-jar", "", "Path to Mustang JAR to verify")
	return cmd
}
