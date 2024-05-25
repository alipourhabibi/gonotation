package notation

import (
	"reflect"
	"testing"
)

func TestBlacklist(t *testing.T) {
	input := map[string]interface{}{
		"name":         "John Doe",
		"avatar":       "avatar.png",
		"email":        "john@example.com",
		"access":       map[string]interface{}{"owner": "admin", "clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}

	exptectedOutput := map[string]interface{}{
		"name":         "John Doe",
		"email":        "john@example.com",
		"organization": "Example Corp",
	}

	blacklistFilters := []string{"*", "!avatar", "!access"}
	filteredBlacklist, err := FilterMap(input, blacklistFilters)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(exptectedOutput, filteredBlacklist) {
		t.Fatalf("Expected %v, got %v", exptectedOutput, filteredBlacklist)
	}
}

func TestWhiteList(t *testing.T) {
	input := map[string]interface{}{
		"name":         "John Doe",
		"avatar":       "avatar.png",
		"email":        "john@example.com",
		"access":       map[string]interface{}{"owner": "admin", "clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}

	exptectedOutput := map[string]interface{}{
		"name":   "John Doe",
		"avatar": "avatar.png",
	}

	whitelistFilters := []string{"name", "avatar"}
	filteredWhitelist, err := FilterMap(input, whitelistFilters)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(exptectedOutput, filteredWhitelist) {
		t.Fatalf("Expected %v, got %v", exptectedOutput, filteredWhitelist)
	}
}

func TestNestedBlacklist(t *testing.T) {
	input := map[string]interface{}{
		"name":         "John Doe",
		"avatar":       "avatar.png",
		"email":        "john@example.com",
		"access":       map[string]interface{}{"owner": "admin", "clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}

	exptectedOutput := map[string]interface{}{
		"name":         "John Doe",
		"email":        "john@example.com",
		"access":       map[string]interface{}{"clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}

	blacklistFilters := []string{"*", "!access.owner", "!avatar"}
	filteredWhitelist, err := FilterMap(input, blacklistFilters)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(exptectedOutput, filteredWhitelist) {
		t.Fatalf("Expected %v, got %v", exptectedOutput, filteredWhitelist)
	}
}

func TestBlacklistNestedWhiltelist(t *testing.T) {
	input := map[string]interface{}{
		"name":         "John Doe",
		"avatar":       "avatar.png",
		"email":        "john@example.com",
		"access":       map[string]interface{}{"owner": "admin", "clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}

	exptectedOutput := map[string]interface{}{
		"name":         "John Doe",
		"email":        "john@example.com",
		"access":       map[string]interface{}{"clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}

	blacklistFilters := []string{"*", "!access", "!avatar", "access.clients"}
	filteredWhitelist, err := FilterMap(input, blacklistFilters)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(exptectedOutput, filteredWhitelist) {
		t.Fatalf("Expected %v, got %v", exptectedOutput, filteredWhitelist)
	}
}

func TestBlacklistWholeField(t *testing.T) {
	input := map[string]interface{}{
		"name":         "John Doe",
		"avatar":       "avatar.png",
		"email":        "john@example.com",
		"access":       map[string]interface{}{"owner": "admin", "clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}

	exptectedOutput := map[string]interface{}{
		"name":         "John Doe",
		"email":        "john@example.com",
		"access":       map[string]interface{}{},
		"organization": "Example Corp",
	}

	blacklistFilters := []string{"*", "!access.owner", "!avatar", "!access.clients"}
	filteredWhitelist, err := FilterMap(input, blacklistFilters)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(exptectedOutput, filteredWhitelist) {
		t.Fatalf("Expected %v, got %v", exptectedOutput, filteredWhitelist)
	}
}
