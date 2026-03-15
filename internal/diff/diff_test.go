package diff

import (
	"testing"

	"github.com/onatbaker/openapi-guard/internal/model"
)

func TestDiffResponseSchema_NestedFieldRemoved(t *testing.T) {
	oldS := model.SchemaObject{
		Fields: map[string]model.Field{
			"address": {
				Type: "object",
				Children: &model.SchemaObject{
					Fields: map[string]model.Field{
						"street": {Type: "string"},
						"city":   {Type: "string"},
					},
					Required: map[string]bool{"street": true},
				},
			},
		},
		Required: map[string]bool{},
	}
	newS := model.SchemaObject{
		Fields: map[string]model.Field{
			"address": {
				Type: "object",
				Children: &model.SchemaObject{
					Fields:   map[string]model.Field{"city": {Type: "string"}},
					Required: map[string]bool{},
				},
			},
		},
		Required: map[string]bool{},
	}

	changes := diffResponseSchema("GET /users/{id}", "", oldS, newS)

	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != FieldRemoved {
		t.Errorf("expected FieldRemoved, got %s", changes[0].Kind)
	}
	if changes[0].FieldName != "address.street" {
		t.Errorf("expected field name 'address.street', got %q", changes[0].FieldName)
	}
	if !changes[0].OldFieldWasRequired {
		t.Error("expected OldFieldWasRequired to be true")
	}
}

func TestDiffResponseSchema_NestedFieldTypeChanged(t *testing.T) {
	oldS := model.SchemaObject{
		Fields: map[string]model.Field{
			"address": {
				Type: "object",
				Children: &model.SchemaObject{
					Fields:   map[string]model.Field{"zip": {Type: "string"}},
					Required: map[string]bool{},
				},
			},
		},
		Required: map[string]bool{},
	}
	newS := model.SchemaObject{
		Fields: map[string]model.Field{
			"address": {
				Type: "object",
				Children: &model.SchemaObject{
					Fields:   map[string]model.Field{"zip": {Type: "integer"}},
					Required: map[string]bool{},
				},
			},
		},
		Required: map[string]bool{},
	}

	changes := diffResponseSchema("GET /users/{id}", "", oldS, newS)

	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != FieldTypeChanged {
		t.Errorf("expected FieldTypeChanged, got %s", changes[0].Kind)
	}
	if changes[0].FieldName != "address.zip" {
		t.Errorf("expected 'address.zip', got %q", changes[0].FieldName)
	}
}

func TestDiffRequestBodySchema_NestedRequiredFieldAdded(t *testing.T) {
	oldS := model.SchemaObject{
		Fields: map[string]model.Field{
			"address": {
				Type: "object",
				Children: &model.SchemaObject{
					Fields:   map[string]model.Field{"city": {Type: "string"}},
					Required: map[string]bool{},
				},
			},
		},
		Required: map[string]bool{},
	}
	newS := model.SchemaObject{
		Fields: map[string]model.Field{
			"address": {
				Type: "object",
				Children: &model.SchemaObject{
					Fields: map[string]model.Field{
						"city":   {Type: "string"},
						"street": {Type: "string"},
					},
					Required: map[string]bool{"street": true},
				},
			},
		},
		Required: map[string]bool{},
	}

	changes := diffRequestBodySchema("POST /users", "", oldS, newS)

	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != RequestBodyFieldAdded {
		t.Errorf("expected RequestBodyFieldAdded, got %s", changes[0].Kind)
	}
	if changes[0].FieldName != "address.street" {
		t.Errorf("expected 'address.street', got %q", changes[0].FieldName)
	}
	if !changes[0].FieldRequired {
		t.Error("expected FieldRequired to be true")
	}
}
