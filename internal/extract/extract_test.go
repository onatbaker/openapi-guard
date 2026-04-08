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

func TestExtractParameters_Ref(t *testing.T) {
	doc := spec.Document{
		"components": map[string]any{
			"parameters": map[string]any{
				"UserId": map[string]any{
					"in":       "path",
					"name":     "id",
					"required": true,
					"schema":   map[string]any{"type": "integer"},
				},
			},
		},
	}
	params := []any{
		map[string]any{"$ref": "#/components/parameters/UserId"},
	}

	out := extractParameters(params, spec.NewResolver(doc))
	p, ok := out[struct{ In, Name string }{"path", "id"}]
	if !ok {
		t.Fatal("expected parameter 'id' in path to be extracted")
	}
	if !p.Required {
		t.Error("expected parameter to be required")
	}
}

func TestExtractResponses_Ref(t *testing.T) {
	doc := spec.Document{
		"components": map[string]any{
			"responses": map[string]any{
				"NotFound": map[string]any{
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"message": map[string]any{"type": "string"},
								},
							},
						},
					},
				},
			},
		},
	}
	op := map[string]any{
		"responses": map[string]any{
			"404": map[string]any{"$ref": "#/components/responses/NotFound"},
		},
	}

	out, err := extractResponses(op, spec.NewResolver(doc))
	if err != nil {
		t.Fatal(err)
	}
	schema, ok := out["404"]
	if !ok {
		t.Fatal("expected 404 response to be extracted")
	}
	if schema.Fields["message"].Type != "string" {
		t.Errorf("expected message type 'string', got %q", schema.Fields["message"].Type)
	}
}

func TestExtractRequestBody_Ref(t *testing.T) {
	doc := spec.Document{
		"components": map[string]any{
			"requestBodies": map[string]any{
				"CreateUser": map[string]any{
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"username": map[string]any{"type": "string"},
								},
								"required": []any{"username"},
							},
						},
					},
				},
			},
		},
	}
	op := map[string]any{
		"requestBody": map[string]any{"$ref": "#/components/requestBodies/CreateUser"},
	}

	body, hasBody, err := extractRequestBody(op, spec.NewResolver(doc))
	if err != nil {
		t.Fatal(err)
	}
	if !hasBody {
		t.Fatal("expected request body to be extracted")
	}
	if body.Fields["username"].Type != "string" {
		t.Errorf("expected username type 'string', got %q", body.Fields["username"].Type)
	}
	if !body.Required["username"] {
		t.Error("expected username to be required")
	}
}

func TestExtractObjectSchema_AllOfTopLevel(t *testing.T) {
	doc := spec.Document{
		"components": map[string]any{
			"schemas": map[string]any{
				"User": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":   map[string]any{"type": "integer"},
						"name": map[string]any{"type": "string"},
					},
					"required": []any{"id"},
				},
			},
		},
	}
	schema := map[string]any{
		"allOf": []any{
			map[string]any{"$ref": "#/components/schemas/User"},
		},
	}

	obj, err := extractObjectSchema(schema, spec.NewResolver(doc))
	if err != nil {
		t.Fatal(err)
	}
	if obj.Fields["id"].Type != "integer" {
		t.Errorf("expected id type 'integer', got %q", obj.Fields["id"].Type)
	}
	if !obj.Required["id"] {
		t.Error("expected 'id' to be required")
	}
}

func TestExtractObjectSchema_AllOfProperty(t *testing.T) {
	doc := spec.Document{
		"components": map[string]any{
			"schemas": map[string]any{
				"Address": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"street": map[string]any{"type": "string"},
						"city":   map[string]any{"type": "string"},
					},
					"required": []any{"street"},
				},
			},
		},
	}
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"address": map[string]any{
				"allOf": []any{
					map[string]any{"$ref": "#/components/schemas/Address"},
				},
			},
		},
	}

	obj, err := extractObjectSchema(schema, spec.NewResolver(doc))
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
		t.Error("expected 'street' to be required")
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
