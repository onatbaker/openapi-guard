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

	changes := diffResponseSchema("GET /users/{id}", "", "200", oldS, newS)

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

	changes := diffResponseSchema("GET /users/{id}", "", "200", oldS, newS)

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

func TestDiffResponses_CodeRemoved(t *testing.T) {
	oldEp := model.Endpoint{
		Responses: map[string]model.SchemaObject{
			"200": {Fields: map[string]model.Field{"id": {Type: "string"}}, Required: map[string]bool{}},
			"422": {Fields: map[string]model.Field{"message": {Type: "string"}}, Required: map[string]bool{}},
		},
	}
	newEp := model.Endpoint{
		Responses: map[string]model.SchemaObject{
			"200": {Fields: map[string]model.Field{"id": {Type: "string"}}, Required: map[string]bool{}},
		},
	}

	changes := diffResponses("POST /users", oldEp, newEp)

	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != ResponseCodeRemoved {
		t.Errorf("expected ResponseCodeRemoved, got %s", changes[0].Kind)
	}
	if changes[0].ResponseCode != "422" {
		t.Errorf("expected ResponseCode '422', got %q", changes[0].ResponseCode)
	}
}

func TestDiffResponses_404FieldRemoved(t *testing.T) {
	oldEp := model.Endpoint{
		Responses: map[string]model.SchemaObject{
			"404": {
				Fields:   map[string]model.Field{"message": {Type: "string"}, "code": {Type: "string"}},
				Required: map[string]bool{},
			},
		},
	}
	newEp := model.Endpoint{
		Responses: map[string]model.SchemaObject{
			"404": {
				Fields:   map[string]model.Field{"message": {Type: "string"}},
				Required: map[string]bool{},
			},
		},
	}

	changes := diffResponses("GET /users/{id}", oldEp, newEp)

	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != FieldRemoved {
		t.Errorf("expected FieldRemoved, got %s", changes[0].Kind)
	}
	if changes[0].ResponseCode != "404" {
		t.Errorf("expected ResponseCode '404', got %q", changes[0].ResponseCode)
	}
	if changes[0].FieldName != "code" {
		t.Errorf("expected FieldName 'code', got %q", changes[0].FieldName)
	}
}
