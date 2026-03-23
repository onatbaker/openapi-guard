package rules

import (
	"strings"

	"github.com/onatbaker/openapi-guard/internal/diff"
)

type Severity string

const (
	Breaking Severity = "BREAKING"
	Warning  Severity = "WARNING"
	Info     Severity = "INFO"
)

type Result struct {
	Severity Severity
	Change   diff.Change
	Message  string
}

type Results struct {
	Items []Result
}

func (r Results) HasBreaking() bool {
	for _, it := range r.Items {
		if it.Severity == Breaking {
			return true
		}
	}
	return false
}

func Classify(changes []diff.Change) Results {
	var out Results

	for _, c := range changes {
		switch c.Kind {
		case diff.EndpointRemoved:
			out.Items = append(out.Items, Result{
				Severity: Breaking,
				Change:   c,
				Message:  c.Endpoint + " removed",
			})

		case diff.ParamAdded:
			if c.ParamRequired {
				out.Items = append(out.Items, Result{
					Severity: Breaking,
					Change:   c,
					Message:  c.Endpoint + " added required " + c.ParamIn + " param '" + c.ParamName + "'",
				})
			} else {
				out.Items = append(out.Items, Result{
					Severity: Info,
					Change:   c,
					Message:  c.Endpoint + " added optional " + c.ParamIn + " param '" + c.ParamName + "'",
				})
			}

		case diff.ParamRemoved:
			out.Items = append(out.Items, Result{
				Severity: Breaking,
				Change:   c,
				Message:  c.Endpoint + " removed " + c.ParamIn + " param '" + c.ParamName + "'",
			})

		case diff.FieldRemoved:
			sev := Breaking
			if !strings.HasPrefix(c.ResponseCode, "2") {
				sev = Warning
			}
			out.Items = append(out.Items, Result{
				Severity: sev,
				Change:   c,
				Message:  c.Endpoint + " removed " + c.ResponseCode + " response field '" + c.FieldName + "'",
			})

		case diff.RequiredFieldDemoted:
			sev := Breaking
			if !strings.HasPrefix(c.ResponseCode, "2") {
				sev = Warning
			}
			out.Items = append(out.Items, Result{
				Severity: sev,
				Change:   c,
				Message:  c.Endpoint + " " + c.ResponseCode + " response field '" + c.FieldName + "' is no longer required",
			})

		case diff.FieldTypeChanged:
			sev := Breaking
			if !strings.HasPrefix(c.ResponseCode, "2") {
				sev = Warning
			}
			out.Items = append(out.Items, Result{
				Severity: sev,
				Change:   c,
				Message:  c.Endpoint + " changed " + c.ResponseCode + " response field '" + c.FieldName + "' type " + c.OldType + " -> " + c.NewType,
			})

		case diff.EnumNarrowed:
			sev := Breaking
			if !strings.HasPrefix(c.ResponseCode, "2") {
				sev = Warning
			}
			out.Items = append(out.Items, Result{
				Severity: sev,
				Change:   c,
				Message:  c.Endpoint + " narrowed enum for " + c.ResponseCode + " response field '" + c.FieldName + "'",
			})

		case diff.ParamTypeChanged:
			out.Items = append(out.Items, Result{
				Severity: Breaking,
				Change:   c,
				Message:  c.Endpoint + " changed " + c.ParamIn + " param '" + c.ParamName + "' type " + c.OldType + " -> " + c.NewType,
			})

		case diff.ParamEnumNarrowed:
			out.Items = append(out.Items, Result{
				Severity: Breaking,
				Change:   c,
				Message:  c.Endpoint + " narrowed enum for " + c.ParamIn + " param '" + c.ParamName + "'",
			})

		case diff.RequestBodyFieldAdded:
			if c.FieldRequired {
				out.Items = append(out.Items, Result{
					Severity: Breaking,
					Change:   c,
					Message:  c.Endpoint + " added required request body field '" + c.FieldName + "'",
				})
			} else {
				out.Items = append(out.Items, Result{
					Severity: Info,
					Change:   c,
					Message:  c.Endpoint + " added optional request body field '" + c.FieldName + "'",
				})
			}

		case diff.RequestBodyFieldRemoved:
			if c.OldFieldWasRequired {
				out.Items = append(out.Items, Result{
					Severity: Warning,
					Change:   c,
					Message:  c.Endpoint + " removed required request body field '" + c.FieldName + "'",
				})
			} else {
				out.Items = append(out.Items, Result{
					Severity: Info,
					Change:   c,
					Message:  c.Endpoint + " removed optional request body field '" + c.FieldName + "'",
				})
			}

		case diff.RequestBodyFieldTypeChanged:
			out.Items = append(out.Items, Result{
				Severity: Breaking,
				Change:   c,
				Message:  c.Endpoint + " changed request body field '" + c.FieldName + "' type " + c.OldType + " -> " + c.NewType,
			})

		case diff.RequestBodyRequiredPromoted:
			out.Items = append(out.Items, Result{
				Severity: Breaking,
				Change:   c,
				Message:  c.Endpoint + " request body field '" + c.FieldName + "' is now required",
			})

		case diff.RequestBodyEnumNarrowed:
			out.Items = append(out.Items, Result{
				Severity: Breaking,
				Change:   c,
				Message:  c.Endpoint + " narrowed enum for request body field '" + c.FieldName + "'",
			})

		case diff.EndpointAdded:
			out.Items = append(out.Items, Result{
				Severity: Info,
				Change:   c,
				Message:  c.Endpoint + " added",
			})

		default:
			out.Items = append(out.Items, Result{
				Severity: Info,
				Change:   c,
				Message:  c.Endpoint + " changed (" + string(c.Kind) + ")",
			})
		}
	}

	return out
}
