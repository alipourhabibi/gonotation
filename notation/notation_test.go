package notation

import (
	"reflect"
	"testing"
)

func TestBlacklist(t *testing.T) {
	input := map[string]any{
		"name":         "John Doe",
		"avatar":       "avatar.png",
		"email":        "john@example.com",
		"access":       map[string]any{"owner": "admin", "clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}

	exptectedOutput := map[string]any{
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
	input := map[string]any{
		"name":         "John Doe",
		"avatar":       "avatar.png",
		"email":        "john@example.com",
		"access":       map[string]any{"owner": "admin", "clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}

	exptectedOutput := map[string]any{
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
	input := map[string]any{
		"name":         "John Doe",
		"avatar":       "avatar.png",
		"email":        "john@example.com",
		"access":       map[string]any{"owner": "admin", "clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}

	exptectedOutput := map[string]any{
		"name":         "John Doe",
		"email":        "john@example.com",
		"access":       map[string]any{"clients": []string{"client1", "client2"}},
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
	input := map[string]any{
		"name":         "John Doe",
		"avatar":       "avatar.png",
		"email":        "john@example.com",
		"access":       map[string]any{"owner": "admin", "clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}

	exptectedOutput := map[string]any{
		"name":         "John Doe",
		"email":        "john@example.com",
		"access":       map[string]any{"clients": []string{"client1", "client2"}},
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
	input := map[string]any{
		"name":         "John Doe",
		"avatar":       "avatar.png",
		"email":        "john@example.com",
		"access":       map[string]any{"owner": "admin", "clients": []string{"client1", "client2"}},
		"organization": "Example Corp",
	}

	exptectedOutput := map[string]any{
		"name":         "John Doe",
		"email":        "john@example.com",
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

func TestSiblingDeepNestedWhitelist(t *testing.T) {
	input := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"x": 1,
				"y": 2,
				"z": 3,
			},
		},
	}
	expected := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"x": 1,
				"y": 2,
			},
		},
	}
	got, err := FilterMap(input, []string{"a.b.x", "a.b.y"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("Expected %v, got %v", expected, got)
	}
}

func TestInputNotMutatedByNestedExclude(t *testing.T) {
	input := map[string]any{
		"access": map[string]any{"owner": "admin", "clients": []string{"c1", "c2"}},
	}
	_, err := FilterMap(input, []string{"*", "!access.owner"})
	if err != nil {
		t.Fatal(err)
	}
	if got := input["access"].(map[string]any)["owner"]; got != "admin" {
		t.Fatalf("input was mutated: access.owner is now %v, want \"admin\"", got)
	}
}

func TestInputNotMutatedByWhitelistedExclude(t *testing.T) {
	input := map[string]any{
		"access": map[string]any{"owner": "admin", "clients": []string{"c1", "c2"}},
	}
	_, err := FilterMap(input, []string{"access", "!access.owner"})
	if err != nil {
		t.Fatal(err)
	}
	if got := input["access"].(map[string]any)["owner"]; got != "admin" {
		t.Fatalf("input was mutated: access.owner is now %v, want \"admin\"", got)
	}
}

func TestBlacklistOnlyImpliesAll(t *testing.T) {
	input := map[string]any{
		"name":   "John Doe",
		"avatar": "avatar.png",
		"email":  "john@example.com",
	}
	expected := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	got, err := FilterMap(input, []string{"!avatar"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("Expected %v, got %v", expected, got)
	}
}

func TestBlacklistOnlyNestedImpliesAll(t *testing.T) {
	input := map[string]any{
		"name":   "John Doe",
		"access": map[string]any{"owner": "admin", "clients": []string{"c1", "c2"}},
	}
	expected := map[string]any{
		"name":   "John Doe",
		"access": map[string]any{"clients": []string{"c1", "c2"}},
	}
	got, err := FilterMap(input, []string{"!access.owner"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("Expected %v, got %v", expected, got)
	}
}
