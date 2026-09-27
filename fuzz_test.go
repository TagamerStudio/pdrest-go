package pdrest

import (
	"bytes"
	"encoding/json"
	"net/url"
	"testing"
	"unicode/utf8"
)

func FuzzNormalizeBaseURL(f *testing.F) {
	for _, seed := range []string{
		"127.0.0.1:17993",
		"https://example.com:443",
		"[::1]:17993",
		"[fe80::1%25eth0]:17993",
		"http://example.com/path",
		"http://user:password@example.com:17993",
		"http://example.com?",
		"http://example.com#frag",
		"",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		normalized, err := normalizeBaseURL(raw, defaultPort)
		if err != nil {
			return
		}
		assertNormalizedBaseURL(t, normalized)
	})
}

func assertNormalizedBaseURL(t *testing.T, normalized string) {
	t.Helper()

	u, err := url.Parse(normalized)
	if err != nil {
		t.Fatalf("accepted URL does not parse: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		t.Fatalf("accepted URL has unsupported scheme: %q", u.Scheme)
	}
	if u.User != nil || u.Path != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		t.Fatalf("accepted URL contains rejected components: %q", normalized)
	}
	if u.Hostname() == "" || !validHostname(u.Hostname()) {
		t.Fatalf("accepted URL has invalid hostname: %q", normalized)
	}
	if _, err := validatePort(u.Port()); err != nil {
		t.Fatalf("accepted URL has invalid port: %v", err)
	}
}

func FuzzGuildStorageUnmarshalJSON(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(`null`),
		[]byte(`{}`),
		[]byte(`{"container_id":"guild-1","current":1,"max":10,"0":{"item_id":"Money","count":5}}`),
		[]byte(`{"0":5}`),
		[]byte(`{"-1":{"item_id":"Money","count":1}}`),
		[]byte(`{"current":"bad"}`),
		[]byte(`[]`),
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var storage GuildStorage
		if err := storage.UnmarshalJSON(data); err != nil {
			return
		}
		for key := range storage.Slots {
			if !isSlotIndex(key) {
				t.Fatalf("accepted non-numeric slot key %q", key)
			}
		}
	})
}

func FuzzForgottenTechsUnmarshalJSON(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(`null`),
		[]byte(`"All"`),
		[]byte(`"all"`),
		[]byte(`"Technology_1"`),
		[]byte(`["Technology_1","Technology_2"]`),
		[]byte(`""`),
		[]byte(`123`),
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var forgotten ForgottenTechs
		if err := forgotten.UnmarshalJSON(data); err != nil {
			return
		}
		if forgotten.All && len(forgotten.IDs) != 0 {
			t.Fatalf("accepted All with explicit IDs: %+v", forgotten)
		}
	})
}

func FuzzNormalizeItemInputs(f *testing.F) {
	for _, seed := range []string{
		`"Money"`,
		`[[ "Money", 2 ]]`,
		`[{"ItemID":"Money","Count":1}]`,
		`[{"item_id":"Money","count":0}]`,
		`[{"item_id":"Money","count":2},{"item_id":"Money","count":3}]`,
		`[["Money","Wood"]]`,
		`[[]]`,
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		var parsed any
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			return
		}
		values, ok := parsed.([]any)
		if !ok {
			return
		}
		inputs := make([]ItemInput, 0, len(values))
		for _, value := range values {
			inputs = append(inputs, value)
		}
		normalized, err := normalizeItemInputs(inputs)
		if err != nil {
			return
		}
		assertNormalizedItemGrants(t, normalized)
	})
}

func assertNormalizedItemGrants(t *testing.T, grants []GiveItem) {
	t.Helper()

	if len(grants) == 0 {
		t.Fatal("successful normalization returned no items")
	}
	seen := make(map[string]bool, len(grants))
	for _, item := range grants {
		if item.ItemID == "" || item.Count <= 0 {
			t.Fatalf("accepted invalid item grant: %+v", item)
		}
		if seen[item.ItemID] {
			t.Fatalf("accepted duplicate item grant for %q", item.ItemID)
		}
		seen[item.ItemID] = true
	}
}

func FuzzResponseSnippet(f *testing.F) {
	for _, seed := range [][]byte{
		nil,
		[]byte("short"),
		bytes.Repeat([]byte("a"), 2048),
		[]byte("truncated \xc3\xa9"),
		{0xff, 0xfe, 0xfd},
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		snippet := responseSnippet(data)
		const maxSnippetLen = 1024 + len("...")
		if len(snippet) > maxSnippetLen {
			t.Fatalf("snippet exceeds the documented limit: %d", len(snippet))
		}
		if !utf8.ValidString(snippet) {
			t.Fatalf("snippet is not valid UTF-8: %q", snippet)
		}
	})
}
