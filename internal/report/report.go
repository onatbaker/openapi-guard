package report

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/onatbaker/openapi-guard/internal/rules"
)

func Print(w io.Writer, results rules.Results, format string) error {
	switch format {
	case "text":
		return printText(w, results)
	case "json":
		return printJSON(w, results)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func printText(w io.Writer, results rules.Results) error {
	items := make([]rules.Result, 0, len(results.Items))
	items = append(items, results.Items...)

	sort.Slice(items, func(i, j int) bool {
		if items[i].Severity != items[j].Severity {
			return severityRank(items[i].Severity) < severityRank(items[j].Severity)
		}
		return items[i].Message < items[j].Message
	})

	for _, it := range items {
		fmt.Fprintf(w, "%s: %s\n", it.Severity, it.Message)
	}

	return nil
}

func printJSON(w io.Writer, results rules.Results) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func severityRank(s rules.Severity) int {
	switch s {
	case rules.Breaking:
		return 0
	case rules.Warning:
		return 1
	default:
		return 2
	}
}
