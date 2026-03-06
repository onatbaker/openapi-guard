package spec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Document map[string]any

func LoadFile(path string) (Document, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return loadJSON(path, b)
	case ".yaml", ".yml":
		return loadYAML(path, b)
	default:
		if doc, jerr := loadJSON(path, b); jerr == nil {
			return doc, nil
		}
		if doc, yerr := loadYAML(path, b); yerr == nil {
			return doc, nil
		}
		return nil, fmt.Errorf("unsupported file extension %q (expected .json/.yaml/.yml)", ext)
	}
}

func loadJSON(path string, b []byte) (Document, error) {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("parse json %s: %w", path, err)
	}

	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("json root must be an object: %s", path)
	}

	return Document(m), nil
}

func loadYAML(path string, b []byte) (Document, error) {
	var v any
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("parse yaml %s: %w", path, err)
	}

	converted, ok := convertYAML(v).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("yaml root must be an object: %s", path)
	}

	return Document(converted), nil
}

func convertYAML(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = convertYAML(val)
		}
		return out

	case map[any]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			out[ks] = convertYAML(val)
		}
		return out

	case []any:
		out := make([]any, 0, len(t))
		for _, item := range t {
			out = append(out, convertYAML(item))
		}
		return out

	default:
		return v
	}
}
