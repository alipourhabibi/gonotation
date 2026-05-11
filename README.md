# GoNotation
This repo is used to filter the fields of an input with a filter array.

## Example:
```go
	input := map[string]interface{}{
		"name":         "John Doe",
		"avatar":       "avatar.png",
		"email":        "john@example.com",
		"access":       map[string]interface{}{"owner": "admin", "clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}

	blacklistFilters := []string{"*", "!access", "!avatar", "access.clients"}
	filteredWhitelist, err := notation.FilterMap(input, blacklistFilters)
	if err != nil {
        return err
	}
```

The output is as follows:
```go
	exptectedOutput := map[string]interface{}{
		"name":         "John Doe",
		"email":        "john@example.com",
		"access":       map[string]interface{}{"clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}
```

In the example above our filters are `["*", "!access", "!avatar", "access.clients"]`.

## Filter syntax

- `"*"` — include every top-level field.
- `"field"` — include `field` (whitelist).
- `"!field"` — exclude `field` (blacklist).
- `"a.b"` / `"!a.b"` — include / exclude a nested field using dot notation.
- A filter list containing only excludes (e.g. `["!a"]`) is treated as if `"*"` were also present, so it means "everything except `a`".

## Limitations

- Dot notation traverses `map[string]any` only. Slices and arrays (e.g. `[]any`, `[]map[string]any`) are returned as-is, so paths like `users.id` over a slice are not currently supported.
