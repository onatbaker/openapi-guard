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

func TestExtractObjectSchema_RequiredFields(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id":    map[string]any{"type": "integer"},
			"email": map[string]any{"type": "string"},
		},
		"required": []any{"id"},
	}

	obj, err := extractObjectSchema(schema, spec.NewResolver(spec.Document{}))
	if err != nil {
		t.Fatal(err)
	}

	if !obj.Required["id"] {
		t.Error("expected 'id' to be required")
	}
	if obj.Required["email"] {
		t.Error("expected 'email' to not be required")
	}
}

func TestExtractObjectSchema_RefProperty(t *testing.T) {
	doc := spec.Document{
		"components": map[string]any{
			"schemas": map[string]any{
				"EmailType": map[string]any{"type": "string"},
			},
		},
	}
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"email": map[string]any{"$ref": "#/components/schemas/EmailType"},
		},
	}

	obj, err := extractObjectSchema(schema, spec.NewResolver(doc))
	if err != nil {
		t.Fatal(err)
	}

	if obj.Fields["email"].Type != "string" {
		t.Errorf("expected email type 'string', got %q", obj.Fields["email"].Type)
	}
}

func TestExtractObjectSchema_NestedObject(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"address": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"street": map[string]any{"type": "string"},
					"zip":    map[string]any{"type": "string"},
				},
				"required": []any{"street"},
			},
		},
	}

	obj, err := extractObjectSchema(schema, spec.NewResolver(spec.Document{}))
	if err != nil {
		t.Fatal(err)
	}

	addr := obj.Fields["address"]
	if addr.Type != "object" {
		t.Errorf("expected address type 'object', got %q", addr.Type)
	}
	if addr.Children == nil {
		t.Fatal("expected address to have children")
	}
	if addr.Children.Fields["street"].Type != "string" {
		t.Errorf("expected street type 'string', got %q", addr.Children.Fields["street"].Type)
	}
	if !addr.Children.Required["street"] {
		t.Error("expected 'street' to be required in nested object")
	}
}

func TestExtractTypeAndEnum_Enum(t *testing.T) {
	schema := map[string]any{
		"type": "string",
		"enum": []any{"active", "inactive", "banned"},
	}

	typ, enum := extractTypeAndEnum(schema)
	if typ != "string" {
		t.Errorf("expected type 'string', got %q", typ)
	}
	if len(enum) != 3 {
		t.Fatalf("expected 3 enum values, got %d", len(enum))
	}
	if enum[0] != "active" || enum[1] != "inactive" || enum[2] != "banned" {
		t.Errorf("unexpected enum values: %v", enum)
	}
}
