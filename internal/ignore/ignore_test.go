package ignore

import (
	"os"
	"testing"

	"github.com/onatbaker/openapi-guard/internal/diff"
	"github.com/onatbaker/openapi-guard/internal/rules"
)

func TestFilter_EndpointOnly(t *testing.T) {
	results := rules.Results{
		Items: []rules.Result{
			{Severity: rules.Breaking, Change: diff.Change{Endpoint: "DELETE /users/{id}"}, Message: "DELETE /users/{id} removed"},
			{Severity: rules.Breaking, Change: diff.Change{Endpoint: "GET /users/{id}"}, Message: "GET /users/{id} removed"},
		},
	}
	entries := []Entry{{Endpoint: "DELETE /users/{id}"}}

	out := Filter(results, entries)
	if len(out.Items) != 1 {
		t.Fatalf("expected 1 item after filter, got %d", len(out.Items))
	}
	if out.Items[0].Change.Endpoint != "GET /users/{id}" {
		t.Errorf("wrong item survived filter: %s", out.Items[0].Change.Endpoint)
	}
}

func TestFilter_WithField(t *testing.T) {
	results := rules.Results{
		Items: []rules.Result{
			{Severity: rules.Breaking, Change: diff.Change{Endpoint: "GET /users/{id}", FieldName: "email"}, Message: "removed email"},
			{Severity: rules.Breaking, Change: diff.Change{Endpoint: "GET /users/{id}", FieldName: "name"}, Message: "removed name"},
		},
	}
	entries := []Entry{{Endpoint: "GET /users/{id}", Field: "email"}}

	out := Filter(results, entries)
	if len(out.Items) != 1 {
		t.Fatalf("expected 1 item after filter, got %d", len(out.Items))
	}
	if out.Items[0].Change.FieldName != "name" {
		t.Errorf("wrong field survived filter: %s", out.Items[0].Change.FieldName)
	}
}

func TestFilter_ParamName(t *testing.T) {
	results := rules.Results{
		Items: []rules.Result{
			{Severity: rules.Breaking, Change: diff.Change{Endpoint: "GET /users/{id}", ParamName: "X-Tenant-ID"}, Message: "removed header param"},
			{Severity: rules.Breaking, Change: diff.Change{Endpoint: "GET /users/{id}", ParamName: "version"}, Message: "removed param"},
		},
	}
	entries := []Entry{{Endpoint: "GET /users/{id}", Field: "X-Tenant-ID"}}

	out := Filter(results, entries)
	if len(out.Items) != 1 {
		t.Fatalf("expected 1 item after filter, got %d", len(out.Items))
	}
	if out.Items[0].Change.ParamName != "version" {
		t.Errorf("wrong param survived filter: %s", out.Items[0].Change.ParamName)
	}
}

func TestFilter_NoEntries(t *testing.T) {
	results := rules.Results{
		Items: []rules.Result{
			{Severity: rules.Breaking, Change: diff.Change{Endpoint: "GET /users/{id}"}, Message: "something"},
		},
	}

	out := Filter(results, nil)
	if len(out.Items) != 1 {
		t.Errorf("expected items to be unchanged with no entries")
	}
}

func TestLoadFile(t *testing.T) {
	f, err := os.CreateTemp("", "ignore-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	f.WriteString("ignore:\n  - endpoint: DELETE /users/{id}\n    reason: deprecated\n")
	f.Close()

	entries, err := LoadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Endpoint != "DELETE /users/{id}" {
		t.Errorf("unexpected endpoint: %s", entries[0].Endpoint)
	}
}
