package monitor

import (
	"testing"
	"time"
)

func TestNewTrendStore_Defaults(t *testing.T) {
	ts := NewTrendStore(0)
	if ts.max != 50 {
		t.Fatalf("expected default max 50, got %d", ts.max)
	}
}

func TestTrendStore_GetUnknownJob(t *testing.T) {
	ts := NewTrendStore(10)
	_, ok := ts.Get("missing")
	if ok {
		t.Fatal("expected false for unknown job")
	}
}

func TestTrendStore_RecordAndGet(t *testing.T) {
	ts := NewTrendStore(10)
	ts.Record("backup", 2*time.Second)
	ts.Record("backup", 4*time.Second)
	ts.Record("backup", 6*time.Second)

	e, ok := ts.Get("backup")
	if !ok {
		t.Fatal("expected entry for backup")
	}
	if e.SampleCount != 3 {
		t.Fatalf("expected 3 samples, got %d", e.SampleCount)
	}
	if e.AvgDrift != 4*time.Second {
		t.Fatalf("expected avg 4s, got %v", e.AvgDrift)
	}
}

func TestTrendStore_CapsAtMaxSamples(t *testing.T) {
	ts := NewTrendStore(3)
	for i := 0; i < 10; i++ {
		ts.Record("job", time.Duration(i)*time.Second)
	}
	e, ok := ts.Get("job")
	if !ok {
		t.Fatal("expected entry")
	}
	if e.SampleCount != 3 {
		t.Fatalf("expected 3 samples after cap, got %d", e.SampleCount)
	}
}

func TestTrendStore_TrendingIncreasing(t *testing.T) {
	ts := NewTrendStore(20)
	for i := 1; i <= 8; i++ {
		ts.Record("job", time.Duration(i)*time.Second)
	}
	e, _ := ts.Get("job")
	if e.Trending != "increasing" {
		t.Fatalf("expected increasing, got %s", e.Trending)
	}
}

func TestTrendStore_TrendingDecreasing(t *testing.T) {
	ts := NewTrendStore(20)
	for i := 8; i >= 1; i-- {
		ts.Record("job", time.Duration(i)*time.Second)
	}
	e, _ := ts.Get("job")
	if e.Trending != "decreasing" {
		t.Fatalf("expected decreasing, got %s", e.Trending)
	}
}

func TestTrendStore_All_ReturnsAllJobs(t *testing.T) {
	ts := NewTrendStore(10)
	ts.Record("a", time.Second)
	ts.Record("b", 2*time.Second)
	entries := ts.All()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestTrendStore_StdDevNonNegative(t *testing.T) {
	ts := NewTrendStore(10)
	ts.Record("x", time.Second)
	ts.Record("x", 3*time.Second)
	ts.Record("x", 5*time.Second)
	e, _ := ts.Get("x")
	if e.StdDev < 0 {
		t.Fatalf("expected non-negative stddev, got %v", e.StdDev)
	}
}
