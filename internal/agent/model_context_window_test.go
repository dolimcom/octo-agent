package agent

import (
	"context"
	"strings"
	"testing"
)

func TestPerModelContextWindowDrivesAgentBudgetsAndGauge(t *testing.T) {
	a := New(&summarizeFake{}, "claude-sonnet-5")
	a.SetModelDeployment("claude-sonnet-5", 32_000, "")

	if got := a.ContextWindow(); got != 32_000 {
		t.Fatalf("ContextWindow = %d, want 32000", got)
	}
	if got := a.compactTriggerTokens(); got != 24_000 {
		t.Errorf("compact trigger = %d, want 24000", got)
	}
	if got := a.compactKeepBudget(); got != 9_600 {
		t.Errorf("compact keep budget = %d, want 9600", got)
	}
	a.History.Append(NewUserMessage("hello"))
	if _, window := a.ContextUsage(); window != 32_000 {
		t.Errorf("context gauge window = %d, want 32000", window)
	}
}

func TestSummarizeUsesLiteInstanceWindowForSameModelName(t *testing.T) {
	primary := &modelRecordingFake{summary: "primary"}
	lite := &msgCountRecordingFake{summary: "lite"}
	a := New(primary, "claude-sonnet-5")
	a.SetModelDeployment("claude-sonnet-5", 64_000, "")
	a.SetLiteModelDeployment(lite, "claude-sonnet-5", 32_000, "")
	msgs := []Message{
		NewUserMessage(strings.Repeat("a", 30_000)),
		NewAssistantMessage(strings.Repeat("b", 30_000)),
		NewUserMessage(strings.Repeat("c", 30_000)),
		NewAssistantMessage(strings.Repeat("d", 30_000)),
		NewUserMessage(strings.Repeat("e", 30_000)),
		NewAssistantMessage(strings.Repeat("f", 30_000)),
	}

	if _, err := a.summarize(context.Background(), msgs, nil); err != nil {
		t.Fatal(err)
	}
	if lite.lastN >= len(msgs)+1 {
		t.Fatalf("lite summarizer received %d messages, want fewer than %d for its 32000-token window", lite.lastN, len(msgs)+1)
	}
}
