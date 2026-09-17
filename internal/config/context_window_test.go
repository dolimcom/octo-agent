package config

import (
	"strings"
	"testing"
)

func TestEndpointModelContextWindowRoundTripAndResolution(t *testing.T) {
	setHome(t)
	cfg := Config{
		Endpoints: []Endpoint{
			{ID: "intranet-a", Provider: "custom", BaseURL: "http://a.example/v1", Protocol: "openai", Models: []EndpointModel{{Model: "claude-sonnet-5", ContextWindow: 32_000}}},
			{ID: "intranet-b", Provider: "custom", BaseURL: "http://b.example/v1", Protocol: "openai", Models: []EndpointModel{{Model: "claude-sonnet-5", ContextWindow: 64_000}}},
		},
		Default: "intranet-a::claude-sonnet-5",
	}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for ref, want := range map[string]int{
		"intranet-a::claude-sonnet-5": 32_000,
		"intranet-b::claude-sonnet-5": 64_000,
	} {
		entry, ok := loaded.EntryByModel(ref)
		if !ok {
			t.Fatalf("EntryByModel(%q) did not resolve", ref)
		}
		if got := entry.EffectiveContextWindow(); got != want {
			t.Errorf("EntryByModel(%q) context window = %d, want %d", ref, got, want)
		}
	}
}

func TestEndpointModelContextWindowValidation(t *testing.T) {
	for _, tc := range []struct {
		value int
		valid bool
	}{
		{value: 0, valid: true},
		{value: MinFallbackContextWindow, valid: true},
		{value: 32, valid: false},
		{value: -1, valid: false},
	} {
		cfg := Config{Endpoints: []Endpoint{{ID: "intranet", Provider: "custom", Models: []EndpointModel{{Model: "model", ContextWindow: tc.value}}}}}
		problems := strings.Join(cfg.Validate(), "\n")
		if tc.valid && strings.Contains(problems, "context_window") {
			t.Errorf("context_window %d unexpectedly rejected: %s", tc.value, problems)
		}
		if !tc.valid && !strings.Contains(problems, "context_window") {
			t.Errorf("context_window %d was not rejected: %s", tc.value, problems)
		}
	}
}

func TestLegacyModelContextWindowMigration(t *testing.T) {
	home := setHome(t)
	writeOcto(t, home, "config.yml", "models:\n  - provider: custom\n    model: local-model\n    base_url: http://localhost:8000/v1\n    protocol: openai\n    context_window: 32000\n")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	entry, ok := cfg.EntryByModel("local-model")
	if !ok || entry.ContextWindow != 32_000 {
		t.Fatalf("legacy context window = (%v, %d), want (true, 32000)", ok, entry.ContextWindow)
	}
}
