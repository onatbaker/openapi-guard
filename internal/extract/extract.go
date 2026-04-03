package extract

import (
	"fmt"
	"strings"

	"github.com/onatbaker/openapi-guard/internal/model"
	"github.com/onatbaker/openapi-guard/internal/spec"
)

var httpMethods = map[string]bool{
	"get":     true,
	"post":    true,
	"put":     true,
	"delete":  true,
	"patch":   true,
	"head":    true,
	"options": true,
}

func Extract(doc spec.Document) (model.API, error) {
	api := model.API{
		Endpoints: make(map[string]model.Endpoint),
	}
	resolver := spec.NewResolver(doc)

	pathsAny, ok := doc["paths"]
	if !ok {
		return api, fmt.Errorf("missing 'paths'")
	}

	paths, ok := pathsAny.(map[string]any)
	if !ok {
		return api, fmt.Errorf("'paths' must be an object")
	}

	for path, pathItemAny := range paths {
		pathItem, ok := pathItemAny.(map[string]any)
		if !ok {
			continue
		}

		pathLevelParams := extractParameters(pathItem["parameters"], resolver)

		for method, opAny := range pathItem {
			methodLower := strings.ToLower(method)
			if !httpMethods[methodLower] {
				continue
			}

			op, ok := opAny.(map[string]any)
			if !ok {
				continue
			}

			ep := model.Endpoint{
				Method: strings.ToUpper(methodLower),
				Path:   path,
				Params: make(map[model.ParamKey]model.Param),
			}

			for k, p := range pathLevelParams {
				ep.Params[k] = p
			}

			opParams := extractParameters(op["parameters"], resolver)
			for k, p := range opParams {
				ep.Params[k] = p
			}

			responses, err := extractResponses(op, resolver)
			if err != nil {
				return api, err
			}
			ep.Responses = responses

			body, hasBody, err := extractRequestBody(op, resolver)
			if err != nil {
				return api, err
			}
			if hasBody {
				ep.RequestBody = body
				ep.HasRequestBody = true
			}

			key := endpointKey(ep.Method, ep.Path)
			api.Endpoints[key] = ep
		}
	}

	return api, nil
}

func endpointKey(method string, path string) string {
	return method + " " + path
}

func extractParameters(paramsAny any, resolver *spec.Resolver) map[model.ParamKey]model.Param {
	out := make(map[model.ParamKey]model.Param)

	list, ok := paramsAny.([]any)
	if !ok {
		return out
	}

	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		if ref, ok := m["$ref"].(string); ok {
			resolved, err := resolver.ResolveComponentObject(ref)
			if err != nil {
				continue
			}
			m = resolved
		}

		inStr, _ := m["in"].(string)
		name, _ := m["name"].(string)

		if inStr != "path" && inStr != "query" && inStr != "header" && inStr != "cookie" {
			continue
		}
		if name == "" {
			continue
		}

		required, _ := m["required"].(bool)

		var typ string
		var enum []string

		if schemaAny, ok := m["schema"].(map[string]any); ok {
			resolved, err := resolver.ResolveSchema(schemaAny)
			if err == nil {
				typ, enum = extractTypeAndEnum(resolved)
			}
			// TODO: this ignores resolve errors please fix!!!!!!!!!!!!!!!!!!!!!!!
		}

		p := model.Param{
			In:       inStr,
			Name:     name,
			Required: required,
			Type:     typ,
			Enum:     enum,
		}

		out[model.ParamKey{In: inStr, Name: name}] = p
	}

	return out
}

func extractResponses(op map[string]any, resolver *spec.Resolver) (map[string]model.SchemaObject, error) {
	out := make(map[string]model.SchemaObject)

	responsesAny, ok := op["responses"]
	if !ok {
		return out, nil
	}
	responses, ok := responsesAny.(map[string]any)
	if !ok {
		return out, nil
	}

	for code, respAny := range responses {
		resp, ok := respAny.(map[string]any)
		if !ok {
			continue
		}

		if ref, ok := resp["$ref"].(string); ok {
			resolved, err := resolver.ResolveComponentObject(ref)
			if err != nil {
				continue
			}
			resp = resolved
		}

		contentAny, ok := resp["content"]
		if !ok {
			continue
		}
		content, ok := contentAny.(map[string]any)
		if !ok {
			continue
		}
		appAny, ok := content["application/json"]
		if !ok {
			continue
		}
		app, ok := appAny.(map[string]any)
		if !ok {
			continue
		}
		schemaAny, ok := app["schema"]
		if !ok {
			continue
		}
		schema, ok := schemaAny.(map[string]any)
		if !ok {
			continue
		}
		resolved, err := resolver.ResolveSchema(schema)
		if err != nil {
			return nil, err
		}
		obj, err := extractObjectSchema(resolved, resolver)
		if err != nil {
			return nil, err
		}
		out[code] = obj
	}

	return out, nil
}

func extractRequestBody(op map[string]any, resolver *spec.Resolver) (model.SchemaObject, bool, error) {
	rbAny, ok := op["requestBody"]
	if !ok {
		return model.SchemaObject{}, false, nil
	}
	rb, ok := rbAny.(map[string]any)
	if !ok {
		return model.SchemaObject{}, false, nil
	}

	if ref, ok := rb["$ref"].(string); ok {
		resolved, err := resolver.ResolveComponentObject(ref)
		if err != nil {
			return model.SchemaObject{}, false, nil
		}
		rb = resolved
	}

	contentAny, ok := rb["content"]
	if !ok {
		return model.SchemaObject{}, false, nil
	}
	content, ok := contentAny.(map[string]any)
	if !ok {
		return model.SchemaObject{}, false, nil
	}

	appAny, ok := content["application/json"]
	if !ok {
		return model.SchemaObject{}, false, nil
	}
	app, ok := appAny.(map[string]any)
	if !ok {
		return model.SchemaObject{}, false, nil
	}

	schemaAny, ok := app["schema"]
	if !ok {
		return model.SchemaObject{}, false, nil
	}
	schema, ok := schemaAny.(map[string]any)
	if !ok {
		return model.SchemaObject{}, false, nil
	}

	resolved, err := resolver.ResolveSchema(schema)
	if err != nil {
		return model.SchemaObject{}, true, err
	}

	obj, err := extractObjectSchema(resolved, resolver)
	if err != nil {
		return model.SchemaObject{}, true, err
	}
	return obj, true, nil
}

func extractObjectSchema(schema map[string]any, resolver *spec.Resolver) (model.SchemaObject, error) {
	out := model.SchemaObject{
		Fields:   make(map[string]model.Field),
		Required: make(map[string]bool),
	}

	if anyOf, ok := schema["anyOf"].([]any); ok {
		for _, v := range anyOf {
			m, ok := v.(map[string]any)
			if !ok {
				continue
			}
			if t, _ := m["type"].(string); t == "object" {
				schema = m
				break
			}
		}
	}

	typ, _ := schema["type"].(string)
	if typ != "object" {
		return out, nil
	}

	if reqAny, ok := schema["required"].([]any); ok {
		for _, r := range reqAny {
			if s, ok := r.(string); ok {
				out.Required[s] = true
			}
		}
	}

	propsAny, ok := schema["properties"].(map[string]any)
	if !ok {
		return out, nil
	}

	for name, pAny := range propsAny {
		pm, ok := pAny.(map[string]any)
		if !ok {
			continue
		}

		pmResolved, err := resolver.ResolveSchema(pm)
		if err != nil {
			return model.SchemaObject{}, err
		}

		t, enum := extractTypeAndEnum(pmResolved)
		f := model.Field{Type: t, Enum: enum}
		if t == "object" {
			child, err := extractObjectSchema(pmResolved, resolver)
			if err != nil {
				return model.SchemaObject{}, err
			}
			f.Children = &child
		}
		out.Fields[name] = f
	}

	return out, nil
}

func extractTypeAndEnum(schema map[string]any) (string, []string) {
	if anyOf, ok := schema["anyOf"].([]any); ok {
		for _, v := range anyOf {
			m, ok := v.(map[string]any)
			if !ok {
				continue
			}
			if t, _ := m["type"].(string); t != "" && t != "null" {
				return extractTypeAndEnum(m)
			}
		}
	}

	typ, _ := schema["type"].(string)

	enumAny, ok := schema["enum"].([]any)
	if !ok {
		return typ, nil
	}

	out := make([]string, 0, len(enumAny))
	for _, v := range enumAny {
		s, ok := v.(string)
		if !ok {
			continue
		}
		out = append(out, s)
	}

	return typ, out
}
