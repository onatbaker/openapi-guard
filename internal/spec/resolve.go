package spec

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Resolver struct {
	sections  map[string]map[string]map[string]any
	baseDir   string
	fileCache map[string]map[string]any
}

func NewResolver(doc Document, baseDir string) *Resolver {
	r := &Resolver{
		sections:  make(map[string]map[string]map[string]any),
		baseDir:   baseDir,
		fileCache: make(map[string]map[string]any),
	}

	components, _ := doc["components"].(map[string]any)
	if components == nil {
		return r
	}

	for section, sectionAny := range components {
		sectionMap, ok := sectionAny.(map[string]any)
		if !ok {
			continue
		}
		r.sections[section] = make(map[string]map[string]any)
		for name, node := range sectionMap {
			m, ok := node.(map[string]any)
			if !ok {
				continue
			}
			r.sections[section][name] = m
		}
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

	if !strings.HasPrefix(ref, "#") {
		return r.resolveFileRef(ref)
	}

	name, ok := parseLocalSchemaRef(ref)
	if !ok {
		return nil, fmt.Errorf("unsupported $ref %q (only #/components/schemas/<Name> is supported)", ref)
	}

	if visited[name] {
		return nil, fmt.Errorf("cyclic $ref detected at %q", name)
	}
	visited[name] = true

	target, ok := r.sections["schemas"][name]
	if !ok {
		return nil, fmt.Errorf("unresolved $ref %q (schema %q not found)", ref, name)
	}

	return r.resolveSchema(target, visited)
}

func (r *Resolver) resolveFileRef(ref string) (map[string]any, error) {
	filePart, fragment, _ := strings.Cut(ref, "#")
	if r.baseDir == "" {
		return nil, fmt.Errorf("cannot resolve external $ref %q: no base directory set", ref)
	}

	absPath := filepath.Join(r.baseDir, filePart)

	doc, ok := r.fileCache[absPath]
	if !ok {
		loaded, err := LoadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("external $ref %q: %w", ref, err)
		}
		doc = map[string]any(loaded)
		r.fileCache[absPath] = doc
	}

	if fragment == "" {
		return map[string]any(doc), nil
	}

	parts := strings.Split(strings.TrimPrefix(fragment, "/"), "/")
	var node any = map[string]any(doc)
	for _, part := range parts {
		m, ok := node.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("external $ref %q: cannot traverse into non-object at %q", ref, part)
		}
		node, ok = m[part]
		if !ok {
			return nil, fmt.Errorf("external $ref %q: key %q not found", ref, part)
		}
	}

	result, ok := node.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("external $ref %q: target is not an object", ref)
	}
	return result, nil
}

// ResolveComponentObject resolves any #/components/<section>/<name> $ref, returning
// the raw object. Used for parameter, response, and requestBody $refs.
func (r *Resolver) ResolveComponentObject(ref string) (map[string]any, error) {
	if !strings.HasPrefix(ref, "#") {
		return r.resolveFileRef(ref)
	}

	trimmed := strings.TrimPrefix(ref, "#/components/")
	if trimmed == ref {
		return nil, fmt.Errorf("unsupported $ref %q", ref)
	}
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return nil, fmt.Errorf("malformed $ref %q", ref)
	}
	section, name := parts[0], parts[1]
	sectionMap, ok := r.sections[section]
	if !ok {
		return nil, fmt.Errorf("unresolved $ref %q (section %q not found in components)", ref, section)
	}
	obj, ok := sectionMap[name]
	if !ok {
		return nil, fmt.Errorf("unresolved $ref %q (%q not found in components/%s)", ref, name, section)
	}
	return obj, nil
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
