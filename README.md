# GoNotation
This repo is used to filter the filed of an input with filter array.

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

The out put is as follow:
```go
	exptectedOutput := map[string]interface{}{
		"name":         "John Doe",
		"email":        "john@example.com",
		"access":       map[string]interface{}{"clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}
```

in the example above our globs are ["*", "!access", "!avatar", "access.clients"] \
