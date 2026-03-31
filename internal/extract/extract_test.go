package extract

import (
	"testing"

	"github.com/onatbaker/openapi-guard/internal/spec"
)

func TestExtractObjectSchema_AnyOfField(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"email": map[string]any{
				"anyOf": []any{
					map[string]any{"type": "string"},
					map[string]any{"type": "null"},
				},
			},
		},
		"required": []any{"name"},
	}

	obj, err := extractObjectSchema(schema, spec.NewResolver(spec.Document{}))
	if err != nil {
		t.Fatal(err)
	}

	if obj.Fields["email"].Type != "string" {
		t.Errorf("expected email type 'string', got %q", obj.Fields["email"].Type)
	}
	if obj.Fields["name"].Type != "string" {
		t.Errorf("expected name type 'string', got %q", obj.Fields["name"].Type)
	}
}
