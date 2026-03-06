package integration

import (
	"strings"
	"testing"

	"github.com/onatbaker/openapi-guard/internal/diff"
	"github.com/onatbaker/openapi-guard/internal/extract"
	"github.com/onatbaker/openapi-guard/internal/rules"
	"github.com/onatbaker/openapi-guard/internal/spec"
)

func TestDiff_FixturesHaveBreakingChanges(t *testing.T) {
	oldDoc, err := spec.LoadFile("../../testdata/old.yaml")
	if err != nil {
		t.Fatalf("load old: %v", err)
	}
	newDoc, err := spec.LoadFile("../../testdata/new.yaml")
	if err != nil {
		t.Fatalf("load new: %v", err)
	}

	oldAPI, err := extract.Extract(oldDoc)
	if err != nil {
		t.Fatalf("extract old: %v", err)
	}
	newAPI, err := extract.Extract(newDoc)
	if err != nil {
		t.Fatalf("extract new: %v", err)
	}

	changes := diff.Diff(oldAPI, newAPI)
	results := rules.Classify(changes)

	if !results.HasBreaking() {
		t.Fatalf("expected breaking changes, got none")
	}

	assertHasBreakingSubstring(t, results, "GET /health removed")
	assertHasBreakingSubstring(t, results, "added required query param 'currency'")
	assertHasBreakingSubstring(t, results, "response field 'id' is no longer required")
}

func assertHasBreakingSubstring(t *testing.T, results rules.Results, substr string) {
	t.Helper()

	for _, it := range results.Items {
		if it.Severity != rules.Breaking {
			continue
		}
		if strings.Contains(it.Message, substr) {
			return
		}
	}

	t.Fatalf("expected breaking message containing %q", substr)
}
