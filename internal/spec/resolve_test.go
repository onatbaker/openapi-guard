package spec

import "testing"

func TestResolveComponentObject_Parameter(t *testing.T) {
	doc := Document{
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

	r := NewResolver(doc)
	obj, err := r.ResolveComponentObject("#/components/parameters/UserId")
	if err != nil {
		t.Fatal(err)
	}
	if obj["name"] != "id" {
		t.Errorf("expected name 'id', got %v", obj["name"])
	}
}

func TestResolveComponentObject_UnknownSection(t *testing.T) {
	r := NewResolver(Document{})
	_, err := r.ResolveComponentObject("#/components/parameters/Foo")
	if err == nil {
		t.Error("expected error for unknown section")
	}
}

func TestResolveSchema_StillWorks(t *testing.T) {
	doc := Document{
		"components": map[string]any{
			"schemas": map[string]any{
				"Name": map[string]any{"type": "string"},
			},
		},
	}

	r := NewResolver(doc)
	resolved, err := r.ResolveSchema(map[string]any{"$ref": "#/components/schemas/Name"})
	if err != nil {
		t.Fatal(err)
	}
	if resolved["type"] != "string" {
		t.Errorf("expected type 'string', got %v", resolved["type"])
	}
}
