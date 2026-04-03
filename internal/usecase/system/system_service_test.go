package system

import "testing"

func TestNormalizeAccessLogRetentionDays(t *testing.T) {
	if got := normalizeAccessLogRetentionDays(0); got != defaultAccessLogRetentionDays {
		t.Fatalf("expected default retention for zero, got %d", got)
	}
	if got := normalizeAccessLogRetentionDays(-5); got != defaultAccessLogRetentionDays {
		t.Fatalf("expected default retention for negative value, got %d", got)
	}
	if got := normalizeAccessLogRetentionDays(45); got != 45 {
		t.Fatalf("expected explicit retention to be preserved, got %d", got)
	}
	if got := normalizeAccessLogRetentionDays(maxAccessLogRetentionDays + 1); got != maxAccessLogRetentionDays {
		t.Fatalf("expected retention to be capped at %d, got %d", maxAccessLogRetentionDays, got)
	}
}
