package diff

import (
	"sort"

	"github.com/onatbaker/openapi-guard/internal/model"
)

type ChangeKind string

const (
	EndpointRemoved      ChangeKind = "EndpointRemoved"
	ParamAdded           ChangeKind = "ParamAdded"
	ParamRemoved         ChangeKind = "ParamRemoved"
	FieldRemoved         ChangeKind = "FieldRemoved"
	FieldTypeChanged     ChangeKind = "FieldTypeChanged"
	RequiredFieldDemoted ChangeKind = "RequiredFieldDemoted"
	EnumNarrowed         ChangeKind = "EnumNarrowed"
	EndpointAdded        ChangeKind = "EndpointAdded"
)

type Change struct {
	Kind     ChangeKind
	Endpoint string
	Detail   string

	ParamIn       string
	ParamName     string
	ParamRequired bool

	FieldName           string
	OldType             string
	NewType             string
	OldEnum             []string
	NewEnum             []string
	OldFieldWasRequired bool
}

func Diff(oldAPI model.API, newAPI model.API) []Change {
	var changes []Change

	for key, oldEp := range oldAPI.Endpoints {
		newEp, ok := newAPI.Endpoints[key]
		if !ok {
			changes = append(changes, Change{
				Kind:     EndpointRemoved,
				Endpoint: key,
				Detail:   "endpoint removed",
			})
			continue
		}

		changes = append(changes, diffParams(key, oldEp, newEp)...)
		changes = append(changes, diffResponse200(key, oldEp, newEp)...)
	}

	for key := range newAPI.Endpoints {
		if _, ok := oldAPI.Endpoints[key]; !ok {
			changes = append(changes, Change{
				Kind:     EndpointAdded,
				Endpoint: key,
				Detail:   "endpoint added",
			})
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Endpoint != changes[j].Endpoint {
			return changes[i].Endpoint < changes[j].Endpoint
		}
		return changes[i].Kind < changes[j].Kind
	})

	return changes
}

func diffParams(endpoint string, oldEp model.Endpoint, newEp model.Endpoint) []Change {
	var out []Change

	for k, oldP := range oldEp.Params {
		if _, ok := newEp.Params[k]; !ok {
			out = append(out, Change{
				Kind:      ParamRemoved,
				Endpoint:  endpoint,
				ParamIn:   oldP.In,
				ParamName: oldP.Name,
				Detail:    "parameter removed",
			})
		}
	}

	for k, newP := range newEp.Params {
		if _, ok := oldEp.Params[k]; !ok {
			out = append(out, Change{
				Kind:          ParamAdded,
				Endpoint:      endpoint,
				ParamIn:       newP.In,
				ParamName:     newP.Name,
				ParamRequired: newP.Required,
				Detail:        "parameter added",
			})
		}
	}

	return out
}

func diffResponse200(endpoint string, oldEp model.Endpoint, newEp model.Endpoint) []Change {
	var out []Change

	if !oldEp.HasResponse200 || !newEp.HasResponse200 {
		return out
	}

	for name := range oldEp.Response200.Fields {
		_, stillExists := newEp.Response200.Fields[name]
		if !stillExists {
			out = append(out, Change{
				Kind:                FieldRemoved,
				Endpoint:            endpoint,
				FieldName:           name,
				OldFieldWasRequired: oldEp.Response200.Required[name],
				Detail:              "response field removed",
			})
			continue
		}
		if oldEp.Response200.Required[name] && !newEp.Response200.Required[name] {
			out = append(out, Change{
				Kind:      RequiredFieldDemoted,
				Endpoint:  endpoint,
				FieldName: name,
				Detail:    "required response field became optional",
			})
		}
	}

	for name, oldF := range oldEp.Response200.Fields {
		newF, ok := newEp.Response200.Fields[name]
		if !ok {
			continue
		}

		if oldF.Type != "" && newF.Type != "" && oldF.Type != newF.Type {
			out = append(out, Change{
				Kind:      FieldTypeChanged,
				Endpoint:  endpoint,
				FieldName: name,
				OldType:   oldF.Type,
				NewType:   newF.Type,
				Detail:    "response field type changed",
			})
			continue
		}

		if len(oldF.Enum) > 0 {
			if enumNarrowed(oldF.Enum, newF.Enum) {
				out = append(out, Change{
					Kind:      EnumNarrowed,
					Endpoint:  endpoint,
					FieldName: name,
					OldEnum:   oldF.Enum,
					NewEnum:   newF.Enum,
					Detail:    "response enum narrowed",
				})
			}
		}
	}

	return out
}

func enumNarrowed(oldEnum []string, newEnum []string) bool {
	if len(oldEnum) == 0 {
		return false
	}

	set := make(map[string]bool, len(newEnum))
	for _, v := range newEnum {
		set[v] = true
	}

	for _, v := range oldEnum {
		if !set[v] {
			return true
		}
	}

	return false
}
