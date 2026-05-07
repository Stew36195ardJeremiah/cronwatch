package monitor

import (
	"testing"
)

func TestNewRoutingStore_DefaultChannel(t *testing.T) {
	s := NewRoutingStore("log")
	if s.Default() != "log" {
		t.Fatalf("expected default 'log', got %q", s.Default())
	}
}

func TestRoutingStore_SetAndResolve(t *testing.T) {
	s := NewRoutingStore("log")
	if err := s.Set("backup", "slack"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := s.Resolve("backup"); got != "slack" {
		t.Fatalf("expected 'slack', got %q", got)
	}
}

func TestRoutingStore_ResolveFallsBackToDefault(t *testing.T) {
	s := NewRoutingStore("pagerduty")
	if got := s.Resolve("unknown-job"); got != "pagerduty" {
		t.Fatalf("expected 'pagerduty', got %q", got)
	}
}

func TestRoutingStore_Remove(t *testing.T) {
	s := NewRoutingStore("log")
	_ = s.Set("myjob", "slack")
	s.Remove("myjob")
	if got := s.Resolve("myjob"); got != "log" {
		t.Fatalf("expected fallback 'log' after remove, got %q", got)
	}
}

func TestRoutingStore_All_ReturnsSnapshot(t *testing.T) {
	s := NewRoutingStore("log")
	_ = s.Set("job-a", "slack")
	_ = s.Set("job-b", "email")
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(all))
	}
}

func TestRoutingStore_Set_EmptyJobReturnsError(t *testing.T) {
	s := NewRoutingStore("log")
	if err := s.Set("", "slack"); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestRoutingStore_Set_EmptyChannelReturnsError(t *testing.T) {
	s := NewRoutingStore("log")
	if err := s.Set("myjob", ""); err == nil {
		t.Fatal("expected error for empty channel")
	}
}

func TestRoutingStore_SetDefault(t *testing.T) {
	s := NewRoutingStore("log")
	s.SetDefault("slack")
	if s.Default() != "slack" {
		t.Fatalf("expected 'slack', got %q", s.Default())
	}
	if got := s.Resolve("any-job"); got != "slack" {
		t.Fatalf("expected resolved default 'slack', got %q", got)
	}
}
