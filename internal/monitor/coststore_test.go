package monitor

import (
	"testing"
	"time"
)

func TestNewCostStore_Defaults(t *testing.T) {
	cs := NewCostStore(0)
	if cs.maxSize != 200 {
		t.Fatalf("expected default maxSize 200, got %d", cs.maxSize)
	}
}

func TestCostStore_SummaryUnknownJob(t *testing.T) {
	cs := NewCostStore(10)
	_, ok := cs.Summary("missing")
	if ok {
		t.Fatal("expected false for unknown job")
	}
}

func TestCostStore_RecordAndSummary(t *testing.T) {
	cs := NewCostStore(10)
	cs.Record("backup", 200*time.Millisecond)
	cs.Record("backup", 400*time.Millisecond)

	s, ok := cs.Summary("backup")
	if !ok {
		t.Fatal("expected summary to exist")
	}
	if s.RunCount != 2 {
		t.Fatalf("expected RunCount 2, got %d", s.RunCount)
	}
	if s.TotalMs != 600 {
		t.Fatalf("expected TotalMs 600, got %d", s.TotalMs)
	}
	if s.AvgMs != 300 {
		t.Fatalf("expected AvgMs 300, got %d", s.AvgMs)
	}
	if s.MaxMs != 400 {
		t.Fatalf("expected MaxMs 400, got %d", s.MaxMs)
	}
}

func TestCostStore_EvictsOldestWhenFull(t *testing.T) {
	cs := NewCostStore(3)
	cs.Record("job", 100*time.Millisecond)
	cs.Record("job", 200*time.Millisecond)
	cs.Record("job", 300*time.Millisecond)
	cs.Record("job", 900*time.Millisecond) // evicts 100ms

	s, _ := cs.Summary("job")
	if s.RunCount != 3 {
		t.Fatalf("expected 3 samples after eviction, got %d", s.RunCount)
	}
	if s.MaxMs != 900 {
		t.Fatalf("expected MaxMs 900, got %d", s.MaxMs)
	}
	// 200+300+900 = 1400
	if s.TotalMs != 1400 {
		t.Fatalf("expected TotalMs 1400, got %d", s.TotalMs)
	}
}

func TestCostStore_All_ReturnsAllJobs(t *testing.T) {
	cs := NewCostStore(10)
	cs.Record("alpha", 50*time.Millisecond)
	cs.Record("beta", 75*time.Millisecond)
	cs.Record("beta", 125*time.Millisecond)

	all := cs.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(all))
	}
	jobs := map[string]CostSummary{}
	for _, s := range all {
		jobs[s.Job] = s
	}
	if jobs["alpha"].RunCount != 1 {
		t.Fatalf("alpha RunCount mismatch")
	}
	if jobs["beta"].RunCount != 2 {
		t.Fatalf("beta RunCount mismatch")
	}
	if jobs["beta"].AvgMs != 100 {
		t.Fatalf("beta AvgMs expected 100, got %d", jobs["beta"].AvgMs)
	}
}

func TestCostStore_MultipleJobs_Independent(t *testing.T) {
	cs := NewCostStore(5)
	cs.Record("jobA", 1*time.Second)
	cs.Record("jobB", 2*time.Second)

	a, _ := cs.Summary("jobA")
	b, _ := cs.Summary("jobB")

	if a.TotalMs != 1000 {
		t.Fatalf("jobA TotalMs expected 1000, got %d", a.TotalMs)
	}
	if b.TotalMs != 2000 {
		t.Fatalf("jobB TotalMs expected 2000, got %d", b.TotalMs)
	}
}
