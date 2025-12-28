package notation

import (
	"encoding/json"
	"strings"
)

type FilterType uint8

const (
	Include FilterType = iota
	Exclude
)

type Filter struct {
	Type  FilterType
	Field string
}

type filterSet struct {
	includeAll    bool
	includes      map[string]bool
	excludes      map[string]bool
	nestedInclude map[string][]string
	nestedExclude map[string][]string
}

func parseFilters(filters []string) []Filter {
	result := make([]Filter, 0, len(filters))
	for _, f := range filters {
		switch {
		case f == "*":
			result = append(result, Filter{Type: Include, Field: "*"})
		case strings.HasPrefix(f, "!"):
			result = append(result, Filter{Type: Exclude, Field: strings.TrimPrefix(f, "!")})
		default:
			result = append(result, Filter{Type: Include, Field: f})
		}
	}
	return result
}

func buildFilterSet(filters []Filter) *filterSet {
	fs := &filterSet{
		includes:      make(map[string]bool),
		excludes:      make(map[string]bool),
		nestedInclude: make(map[string][]string),
		nestedExclude: make(map[string][]string),
	}

	for _, filter := range filters {
		if filter.Type == Include && filter.Field == "*" {
			fs.includeAll = true
			continue
		}

		parent, sub, isNested := splitField(filter.Field)

		if filter.Type == Include {
			if isNested {
				fs.nestedInclude[parent] = append(fs.nestedInclude[parent], sub)
			} else {
				fs.includes[filter.Field] = true
			}
		} else {
			if isNested {
				fs.nestedExclude[parent] = append(fs.nestedExclude[parent], sub)
			} else {
				fs.excludes[filter.Field] = true
			}
		}
	}

	return fs
}

func splitField(field string) (parent, sub string, isNested bool) {
	idx := strings.Index(field, ".")
	if idx == -1 {
		return field, "", false
	}
	return field[:idx], field[idx+1:], true
}

func applyFilters(input map[string]any, filters []Filter) map[string]any {
	fs := buildFilterSet(filters)
	result := make(map[string]any)

	// Start with included fields
	if fs.includeAll {
		for k, v := range input {
			if !fs.excludes[k] {
				result[k] = v
			}
		}
	} else if len(fs.includes) > 0 {
		for field := range fs.includes {
			if v, ok := input[field]; ok {
				result[field] = v
			}
		}
	}

	// Apply top-level excludes
	for field := range fs.excludes {
		delete(result, field)
	}

	// Handle nested includes
	for parent, subFields := range fs.nestedInclude {
		if val, ok := input[parent]; ok {
			if nested, ok := val.(map[string]any); ok {
				filtered := filterNested(nested, subFields, true)
				if len(filtered) > 0 {
					if existing, exists := result[parent].(map[string]any); exists {
						// Merge with existing
						for k, v := range filtered {
							existing[k] = v
						}
					} else {
						result[parent] = filtered
					}
				}
			}
		}
	}

	// Handle nested excludes
	for parent, subFields := range fs.nestedExclude {
		if resultVal, ok := result[parent].(map[string]any); ok {
			for _, subField := range subFields {
				deleteNested(resultVal, subField)
			}
			if len(resultVal) == 0 {
				delete(result, parent)
			}
		}
	}

	return result
}

func filterNested(input map[string]any, fields []string, includeMode bool) map[string]any {
	result := make(map[string]any)

	for _, field := range fields {
		parent, sub, isNested := splitField(field)

		if !isNested {
			if v, ok := input[parent]; ok {
				result[parent] = v
			}
		} else {
			if val, ok := input[parent]; ok {
				if nested, ok := val.(map[string]any); ok {
					subResult := filterNested(nested, []string{sub}, includeMode)
					if len(subResult) > 0 {
						result[parent] = subResult
					}
				}
			}
		}
	}

	return result
}

func deleteNested(input map[string]any, field string) {
	parent, sub, isNested := splitField(field)

	if !isNested {
		delete(input, parent)
	} else {
		if val, ok := input[parent].(map[string]any); ok {
			deleteNested(val, sub)
			if len(val) == 0 {
				delete(input, parent)
			}
		}
	}
}

func toMap(input any) (map[string]any, error) {
	if m, ok := input.(map[string]any); ok {
		return m, nil
	}

	var data []byte
	var err error

	switch v := input.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func FilterMap(input any, filterList []string) (map[string]any, error) {
	inputMap, err := toMap(input)
	if err != nil {
		return nil, err
	}

	filters := parseFilters(filterList)
	return applyFilters(inputMap, filters), nil
}
