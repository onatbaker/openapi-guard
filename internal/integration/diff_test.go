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
	assertHasBreakingSubstring(t, results, "removed header param 'X-Tenant-ID'")
	assertHasBreakingSubstring(t, results, "added required request body field 'role'")
	assertHasBreakingSubstring(t, results, "response field 'address.street'")
	assertHasWarningSubstring(t, results, "404 response field 'code'")
	assertHasWarningSubstring(t, results, "removed 422 response")
}

func TestDiff_NonbreakingChangesAreNotBreaking(t *testing.T) {
	oldDoc, err := spec.LoadFile("../../testdata/old.yaml")
	if err != nil {
		t.Fatalf("load old: %v", err)
	}
	newDoc, err := spec.LoadFile("../../testdata/new_nonbreaking.yaml")
	if err != nil {
		t.Fatalf("load new_nonbreaking: %v", err)
	}

	oldAPI, err := extract.Extract(oldDoc)
	if err != nil {
		t.Fatalf("extract old: %v", err)
	}
	newAPI, err := extract.Extract(newDoc)
	if err != nil {
		t.Fatalf("extract new_nonbreaking: %v", err)
	}

	changes := diff.Diff(oldAPI, newAPI)
	results := rules.Classify(changes)

	if results.HasBreaking() {
		t.Log("unexpected breaking changes:")
		for _, it := range results.Items {
			if it.Severity == rules.Breaking {
				t.Log(" ", it.Message)
			}
		}
		t.Fatalf("expected no breaking changes between old and new_nonbreaking")
	}
}

func TestDiff_FastAPIFixtures(t *testing.T) {
	oldDoc, err := spec.LoadFile("../../testdata/fastapi_old.yaml")
	if err != nil {
		t.Fatalf("load fastapi_old: %v", err)
	}
	newDoc, err := spec.LoadFile("../../testdata/fastapi_new.yaml")
	if err != nil {
		t.Fatalf("load fastapi_new: %v", err)
	}

	oldAPI, err := extract.Extract(oldDoc)
	if err != nil {
		t.Fatalf("extract fastapi_old: %v", err)
	}
	newAPI, err := extract.Extract(newDoc)
	if err != nil {
		t.Fatalf("extract fastapi_new: %v", err)
	}

	results := rules.Classify(diff.Diff(oldAPI, newAPI))

	if !results.HasBreaking() {
		t.Fatal("expected breaking changes in fastapi fixtures")
	}
	assertHasBreakingSubstring(t, results, "removed 200 response field 'email'")
	assertHasBreakingSubstring(t, results, "narrowed enum for 200 response field 'role'")
}

func TestDiff_DjangoFixtures(t *testing.T) {
	oldDoc, err := spec.LoadFile("../../testdata/django_old.yaml")
	if err != nil {
		t.Fatalf("load django_old: %v", err)
	}
	newDoc, err := spec.LoadFile("../../testdata/django_new.yaml")
	if err != nil {
		t.Fatalf("load django_new: %v", err)
	}

	oldAPI, err := extract.Extract(oldDoc)
	if err != nil {
		t.Fatalf("extract django_old: %v", err)
	}
	newAPI, err := extract.Extract(newDoc)
	if err != nil {
		t.Fatalf("extract django_new: %v", err)
	}

	results := rules.Classify(diff.Diff(oldAPI, newAPI))

	if !results.HasBreaking() {
		t.Fatal("expected breaking changes in django fixtures")
	}
	assertHasBreakingSubstring(t, results, "removed 200 response field 'email'")
	assertHasBreakingSubstring(t, results, "narrowed enum for 200 response field 'status'")
}

func TestDiff_ASPNETFixtures(t *testing.T) {
	oldDoc, err := spec.LoadFile("../../testdata/aspnet_old.yaml")
	if err != nil {
		t.Fatalf("load aspnet_old: %v", err)
	}
	newDoc, err := spec.LoadFile("../../testdata/aspnet_new.yaml")
	if err != nil {
		t.Fatalf("load aspnet_new: %v", err)
	}

	oldAPI, err := extract.Extract(oldDoc)
	if err != nil {
		t.Fatalf("extract aspnet_old: %v", err)
	}
	newAPI, err := extract.Extract(newDoc)
	if err != nil {
		t.Fatalf("extract aspnet_new: %v", err)
	}

	results := rules.Classify(diff.Diff(oldAPI, newAPI))

	if !results.HasBreaking() {
		t.Fatal("expected breaking changes in aspnet fixtures")
	}
	assertHasBreakingSubstring(t, results, "removed header param 'X-Api-Version'")
	assertHasBreakingSubstring(t, results, "removed 200 response field 'description'")
	assertHasBreakingSubstring(t, results, "request body field 'category' is now required")
}

func TestDiff_NestJSFixtures(t *testing.T) {
	oldDoc, err := spec.LoadFile("../../testdata/nestjs_old.yaml")
	if err != nil {
		t.Fatalf("load nestjs_old: %v", err)
	}
	newDoc, err := spec.LoadFile("../../testdata/nestjs_new.yaml")
	if err != nil {
		t.Fatalf("load nestjs_new: %v", err)
	}

	oldAPI, err := extract.Extract(oldDoc)
	if err != nil {
		t.Fatalf("extract nestjs_old: %v", err)
	}
	newAPI, err := extract.Extract(newDoc)
	if err != nil {
		t.Fatalf("extract nestjs_new: %v", err)
	}

	results := rules.Classify(diff.Diff(oldAPI, newAPI))

	if !results.HasBreaking() {
		t.Fatal("expected breaking changes in nestjs fixtures")
	}
	assertHasBreakingSubstring(t, results, "removed 200 response field 'notes'")
	assertHasBreakingSubstring(t, results, "narrowed enum for 200 response field 'status'")
}

func assertHasWarningSubstring(t *testing.T, results rules.Results, substr string) {
	t.Helper()

	for _, it := range results.Items {
		if it.Severity != rules.Warning {
			continue
		}
		if strings.Contains(it.Message, substr) {
			return
		}
	}

	t.Fatalf("expected warning message containing %q", substr)
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
