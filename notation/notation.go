package notation

import (
	"encoding/json"
	"strings"
)

type FilterType uint8

const (
	Include FilterType = 0
	Exclude FilterType = 1
)

// Filter represents a single filter rule.
type Filter struct {
	Type  FilterType // Include for whitelist, Exclude for blacklist
	Field string     // the field name to include or exclude
}

// parseFilters parses a list of filter strings into a slice of Filter structs.
func parseFilters(filters []string) []Filter {
	var parsedFilters []Filter
	for _, f := range filters {
		if f == "*" {
			parsedFilters = append(parsedFilters, Filter{Type: Include, Field: "*"})
		} else if strings.HasPrefix(f, "!") {
			parsedFilters = append(parsedFilters, Filter{Type: Exclude, Field: strings.TrimPrefix(f, "!")})
		} else {
			parsedFilters = append(parsedFilters, Filter{Type: Include, Field: f})
		}
	}
	return parsedFilters
}

// applyFilters applies the parsed filters to the input map.
func applyFilters(input map[string]interface{}, filters []Filter) map[string]interface{} {
	includeAll := false
	includeFilters := make(map[string]struct{})
	excludeFilters := make(map[string]struct{})
	includeSubFields := make(map[string][]string)
	excludeSubFields := make(map[string][]string)

	for _, filter := range filters {
		if filter.Type == Include && filter.Field == "*" {
			includeAll = true
		} else if filter.Type == Include {
			if strings.Contains(filter.Field, ".") {
				parts := strings.Split(filter.Field, ".")
				parentField := parts[0]
				subField := strings.Join(parts[1:], ".")
				includeSubFields[parentField] = append(includeSubFields[parentField], subField)
			} else {
				includeFilters[filter.Field] = struct{}{}
			}
		} else if filter.Type == Exclude {
			if strings.Contains(filter.Field, ".") {
				parts := strings.Split(filter.Field, ".")
				parentField := parts[0]
				subField := strings.Join(parts[1:], ".")
				excludeSubFields[parentField] = append(excludeSubFields[parentField], subField)
			} else {
				excludeFilters[filter.Field] = struct{}{}
			}
		}
	}

	result := make(map[string]interface{})

	// If there are include filters, use them as a whitelist
	if len(includeFilters) > 0 || includeAll {
		if includeAll {
			for k, v := range input {
				if _, ok := excludeFilters[k]; !ok {
					result[k] = v
				}
			}
		} else {
			for field := range includeFilters {
				copyField(result, input, field)
			}
		}
	}

	// Apply exclude filters
	for field := range excludeFilters {
		removeField(result, input, field)
	}

	// Handle nested includes within nested included fields
	for field, subFields := range includeSubFields {
		if val, ok := input[field]; ok {
			if nestedMap, ok := val.(map[string]interface{}); ok {
				if _, ok := result[field]; !ok {
					result[field] = make(map[string]interface{})
				}
				nestedResult := filterNestedFields(nestedMap, subFields)
				for subKey, subVal := range nestedResult {
					result[field].(map[string]interface{})[subKey] = subVal
				}
			}
		}
	}

	for field, subFields := range excludeSubFields {
		if val, ok := input[field]; ok {
			if nestedMap, ok := val.(map[string]interface{}); ok {
				if resultField, ok := result[field].(map[string]any); ok {
					for _, v := range subFields {
						removeField(resultField, nestedMap, v)
					}
				}
			}
		}
	}

	return result
}

// copyField copies a field from src to dest, supporting nested fields using dot notation.
func copyField(dest, src map[string]interface{}, field string) {
	parts := strings.Split(field, ".")
	if len(parts) == 1 {
		if val, ok := src[parts[0]]; ok {
			dest[parts[0]] = val
		}
	} else {
		if val, ok := src[parts[0]]; ok {
			if nestedMap, ok := val.(map[string]interface{}); ok {
				if _, ok := dest[parts[0]]; !ok {
					dest[parts[0]] = make(map[string]interface{})
				}
				copyField(dest[parts[0]].(map[string]interface{}), nestedMap, strings.Join(parts[1:], "."))
			}
		}
	}
}

// removeField removes a field from the result map if exists in input, supporting nested fields using dot notation.
func removeField(result, input map[string]interface{}, field string) {
	parts := strings.Split(field, ".")
	if len(parts) == 1 {
		delete(result, parts[0])
	} else {
		if val, ok := input[parts[0]]; ok {
			if nestedMap, ok := val.(map[string]interface{}); ok {
				removeField(nestedMap, nestedMap, strings.Join(parts[1:], "."))
				if len(nestedMap) == 0 {
					delete(result, parts[0])
				} else {
					result[parts[0]] = nestedMap
				}
			}
		}
	}
}

// filterNestedFields applies include filters to nested maps.
func filterNestedFields(input map[string]interface{}, fields []string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, field := range fields {
		copyField(result, input, field)
	}
	return result
}

// FilterMap applies the filter list to the input map and returns the filtered map.
func FilterMap(input interface{}, filterList []string) (map[string]interface{}, error) {
	filters := parseFilters(filterList)
	inputMap := map[string]interface{}{}
	switch i := input.(type) {
	case map[string]interface{}:
		inputMap = i
	case string:
		err := json.Unmarshal([]byte(i), &inputMap)
		if err != nil {
			return nil, err
		}
	case []byte:
		err := json.Unmarshal(i, &inputMap)
		if err != nil {
			return nil, err
		}
	default:
		inputBytes, err := json.Marshal(i)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(inputBytes, &inputMap)
		if err != nil {
			return nil, err
		}
	}
	return applyFilters(inputMap, filters), nil
}
