package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 || hasFlag(args, "--help") || hasFlag(args, "-h") {
		printUsage()
		return
	}

	dbPath := extractFlag(args, "-db", "")
	if dbPath == "" {
		fmt.Fprintln(os.Stderr, "ERROR: -db is required")
		printUsage()
		os.Exit(1)
	}

	fromStr := extractFlag(args, "-from", defaultFromDate())
	toStr := extractFlag(args, "-to", defaultToDate())
	worker := parseInt(extractFlag(args, "-workers", "10"))

	startDate, err := parseDate(fromStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: invalid -from '%s', expected YYYY-MM-DD\n", fromStr)
		os.Exit(1)
	}
	endDate, err := parseDate(toStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: invalid -to '%s', expected YYYY-MM-DD\n", toStr)
		os.Exit(1)
	}

	db, err := openDB(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot open database %s: %v\n", dbPath, err)
		os.Exit(1)
	}
	defer db.close()

	updater := NewUpdater(db, dbPath)
	updater.worker = worker

	if hasFlag(args, "-update-concepts") {
		if err := updater.UpdateConcepts(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: concept update failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if startDate > endDate {
		fmt.Fprintln(os.Stderr, "ERROR: -from must be before or equal to -to")
		os.Exit(1)
	}

	if err := updater.UpdateRange(startDate, endDate); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: update failed: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Stock Data Updater - 股票日交易数据更新工具

Usage:
  update.exe -db PATH [-from YYYY-MM-DD] [-to YYYY-MM-DD] [-workers N]
  update.exe -db PATH -update-concepts [-workers N]

Required:
  -db PATH         Database path (SQLite, same schema as jia-stk.db)

Options:
  -update-concepts Update concept tags from Sina Finance (concept boards + constituent stocks)
  -from YYYY-MM-DD  Start date (default: 30 days before today)
  -to YYYY-MM-DD    End date (default: today)
  -workers N        Worker count for backfill/concepts (default: 10)
  --help, -h        Show this help

Examples:
  update.exe -db jia-stk.db
  update.exe -db jia-stk.db -update-concepts
  update.exe -db test.db -from 2025-01-01 -to 2026-07-01
  update.exe -db jia-stk.db -from 2026-07-04 -to 2026-07-04`)
}

func defaultFromDate() string {
	return time.Now().AddDate(0, -1, 0).Format("2006-01-02")
}

func defaultToDate() string {
	return time.Now().Format("2006-01-02")
}

func parseDate(s string) (string, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return "", err
	}
	return t.Format("20060102"), nil
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func extractFlag(args []string, flag, def string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return def
}

func parseInt(s string) int {
	var v int
	fmt.Sscanf(s, "%d", &v)
	if v == 0 {
		return 10
	}
	return v
}

func logSection(title string) {
	n := 60 - len(title)
	if n < 2 {
		n = 2
	}
	pad := strings.Repeat("=", n/2)
	fmt.Printf("\n%s %s %s\n", pad, title, pad)
}

func logInfo(format string, args ...interface{}) {
	fmt.Printf("  %s\n", fmt.Sprintf(format, args...))
}

func logOK(format string, args ...interface{}) {
	fmt.Printf("  [OK] %s\n", fmt.Sprintf(format, args...))
}

func logWarn(format string, args ...interface{}) {
	fmt.Printf("  [WARN] %s\n", fmt.Sprintf(format, args...))
}

func logDebug(format string, args ...interface{}) {
	fmt.Printf("    %s\n", fmt.Sprintf(format, args...))
}
