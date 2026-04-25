package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/onatbaker/openapi-guard/internal/rules"
)

const (
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorReset  = "\033[0m"
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

	color := isTerminal(w)
	for _, it := range items {
		if color {
			fmt.Fprintf(w, "%s%s%s: %s\n", severityColor(it.Severity), it.Severity, colorReset, it.Message)
		} else {
			fmt.Fprintf(w, "%s: %s\n", it.Severity, it.Message)
		}
	}

	return nil
}

func severityColor(s rules.Severity) string {
	switch s {
	case rules.Breaking:
		return colorRed
	case rules.Warning:
		return colorYellow
	default:
		return colorCyan
	}
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
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
