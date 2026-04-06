package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/onatbaker/openapi-guard/internal/diff"
	"github.com/onatbaker/openapi-guard/internal/extract"
	"github.com/onatbaker/openapi-guard/internal/ignore"
	"github.com/onatbaker/openapi-guard/internal/report"
	"github.com/onatbaker/openapi-guard/internal/rules"
	"github.com/onatbaker/openapi-guard/internal/spec"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "diff":
		os.Exit(runDiff(os.Args[2:]))
	case "-h", "--help", "help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}

func runDiff(args []string) int {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	oldPath := fs.String("old", "", "path to old OpenAPI spec (yaml/json)")
	newPath := fs.String("new", "", "path to new OpenAPI spec (yaml/json)")
	format := fs.String("format", "text", "output format: text|json")
	ignorePath := fs.String("ignore", "", "path to ignore config file (yaml)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *oldPath == "" || *newPath == "" {
		fmt.Fprintln(os.Stderr, "missing required flags: --old and --new")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "example: openapi-guard diff --old old.yaml --new new.yaml")
		fmt.Fprintln(os.Stderr)
		return 2
	}

	oldDoc, err := spec.LoadFile(*oldPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load old spec: %v\n", err)
		return 2
	}

	newDoc, err := spec.LoadFile(*newPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load new spec: %v\n", err)
		return 2
	}

	oldAPI, err := extract.Extract(oldDoc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to extract old spec: %v\n", err)
		return 2
	}

	newAPI, err := extract.Extract(newDoc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to extract new spec: %v\n", err)
		return 2
	}

	changes := diff.Diff(oldAPI, newAPI)
	results := rules.Classify(changes)

	if *ignorePath != "" {
		entries, err := ignore.LoadFile(*ignorePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load ignore file: %v\n", err)
			return 2
		}
		results = ignore.Filter(results, entries)
	}

	if err := report.Print(os.Stdout, results, *format); err != nil {
		fmt.Fprintf(os.Stderr, "failed to print report: %v\n", err)
		return 2
	}

	if results.HasBreaking() {
		return 1
	}

	return 0
}

func printUsage() {
	fmt.Println("openapi-guard - OpenAPI contract breakage detector")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  openapi-guard diff --old <file> --new <file> [--format text|json]")
	fmt.Println()
}
