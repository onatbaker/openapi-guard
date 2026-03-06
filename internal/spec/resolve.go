package spec

import (
	"fmt"
	"strings"
)

type Resolver struct {
	schemas map[string]map[string]any
}

func NewResolver(doc Document) *Resolver {
	r := &Resolver{
		schemas: make(map[string]map[string]any),
	}

	components, _ := doc["components"].(map[string]any)
	if components == nil {
		return r
	}

	schemasAny, _ := components["schemas"].(map[string]any)
	if schemasAny == nil {
		return r
	}

	for name, node := range schemasAny {
		m, ok := node.(map[string]any)
		if !ok {
			continue
		}
		r.schemas[name] = m
	}

	return r
}

func (r *Resolver) ResolveSchema(schema map[string]any) (map[string]any, error) {
	visited := make(map[string]bool)
	return r.resolveSchema(schema, visited)
}

func (r *Resolver) resolveSchema(schema map[string]any, visited map[string]bool) (map[string]any, error) {
	ref, _ := schema["$ref"].(string)
	if ref == "" {
		return schema, nil
	}

	name, ok := parseLocalSchemaRef(ref)
	if !ok {
		return nil, fmt.Errorf("unsupported $ref %q (only #/components/schemas/<Name> is supported)", ref)
	}

	if visited[name] {
		return nil, fmt.Errorf("cyclic $ref detected at %q", name)
	}
	visited[name] = true

	target, ok := r.schemas[name]
	if !ok {
		return nil, fmt.Errorf("unresolved $ref %q (schema %q not found)", ref, name)
	}

	return r.resolveSchema(target, visited)
}

func parseLocalSchemaRef(ref string) (string, bool) {
	const prefix = "#/components/schemas/"
	if !strings.HasPrefix(ref, prefix) {
		return "", false
	}
	name := strings.TrimPrefix(ref, prefix)
	if name == "" {
		return "", false
	}
	return name, true
}
