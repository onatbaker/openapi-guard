package rules

import (
	"testing"

	"github.com/onatbaker/openapi-guard/internal/diff"
)

func TestClassify_FieldRemovedSeverity(t *testing.T) {
	breaking := Classify([]diff.Change{{Kind: diff.FieldRemoved, ResponseCode: "200"}})
	if breaking.Items[0].Severity != Breaking {
		t.Errorf("200 field removed should be BREAKING, got %s", breaking.Items[0].Severity)
	}

	warning := Classify([]diff.Change{{Kind: diff.FieldRemoved, ResponseCode: "404"}})
	if warning.Items[0].Severity != Warning {
		t.Errorf("404 field removed should be WARNING, got %s", warning.Items[0].Severity)
	}
}

func TestClassify_ResponseCodeRemovedSeverity(t *testing.T) {
	breaking := Classify([]diff.Change{{Kind: diff.ResponseCodeRemoved, ResponseCode: "201"}})
	if breaking.Items[0].Severity != Breaking {
		t.Errorf("201 code removed should be BREAKING, got %s", breaking.Items[0].Severity)
	}

	warning := Classify([]diff.Change{{Kind: diff.ResponseCodeRemoved, ResponseCode: "422"}})
	if warning.Items[0].Severity != Warning {
		t.Errorf("422 code removed should be WARNING, got %s", warning.Items[0].Severity)
	}
}

func TestClassify_ParamAdded(t *testing.T) {
	breaking := Classify([]diff.Change{{Kind: diff.ParamAdded, ParamRequired: true}})
	if breaking.Items[0].Severity != Breaking {
		t.Errorf("required param added should be BREAKING, got %s", breaking.Items[0].Severity)
	}

	info := Classify([]diff.Change{{Kind: diff.ParamAdded, ParamRequired: false}})
	if info.Items[0].Severity != Info {
		t.Errorf("optional param added should be INFO, got %s", info.Items[0].Severity)
	}
}

func TestClassify_RequestBodyFieldAdded(t *testing.T) {
	breaking := Classify([]diff.Change{{Kind: diff.RequestBodyFieldAdded, FieldRequired: true}})
	if breaking.Items[0].Severity != Breaking {
		t.Errorf("required body field added should be BREAKING, got %s", breaking.Items[0].Severity)
	}

	info := Classify([]diff.Change{{Kind: diff.RequestBodyFieldAdded, FieldRequired: false}})
	if info.Items[0].Severity != Info {
		t.Errorf("optional body field added should be INFO, got %s", info.Items[0].Severity)
	}
}

func TestClassify_RequestBodyFieldRemoved(t *testing.T) {
	warning := Classify([]diff.Change{{Kind: diff.RequestBodyFieldRemoved, OldFieldWasRequired: true}})
	if warning.Items[0].Severity != Warning {
		t.Errorf("required body field removed should be WARNING, got %s", warning.Items[0].Severity)
	}

	info := Classify([]diff.Change{{Kind: diff.RequestBodyFieldRemoved, OldFieldWasRequired: false}})
	if info.Items[0].Severity != Info {
		t.Errorf("optional body field removed should be INFO, got %s", info.Items[0].Severity)
	}
}
