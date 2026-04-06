package ignore

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/onatbaker/openapi-guard/internal/rules"
)

type Entry struct {
	Endpoint string `yaml:"endpoint"`
	Field    string `yaml:"field"`
	Reason   string `yaml:"reason"`
}

type Config struct {
	Ignore []Entry `yaml:"ignore"`
}

func LoadFile(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg.Ignore, nil
}

func Filter(results rules.Results, entries []Entry) rules.Results {
	if len(entries) == 0 {
		return results
	}
	var out rules.Results
	for _, r := range results.Items {
		if !matches(r, entries) {
			out.Items = append(out.Items, r)
		}
	}
	return out
}

func matches(r rules.Result, entries []Entry) bool {
	for _, e := range entries {
		if e.Endpoint != r.Change.Endpoint {
			continue
		}
		if e.Field == "" {
			return true
		}
		if e.Field == r.Change.FieldName || e.Field == r.Change.ParamName {
			return true
		}
	}
	return false
}
