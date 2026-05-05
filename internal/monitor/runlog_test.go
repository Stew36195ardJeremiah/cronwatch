package monitor

import (
	"testing"
	"time"
)

func makeEntry(job string, success bool) RunLogEntry {
	return RunLogEntry{
		JobName:   job,
		StartedAt: time.Now(),
		Duration:  100 * time.Millisecond,
		Success:   success,
		Message:   "",
	}
}

func TestNewRunLog_DefaultMaxSize(t *testing.T) {
	rl := NewRunLog(0)
	if rl.maxSize != 200 {
		t.Errorf("expected default max size 200, got %d", rl.maxSize)
	}
}

func TestRunLog_AppendAndAll(t *testing.T) {
	rl := NewRunLog(10)
	rl.Append(makeEntry("job-a", true))
	rl.Append(makeEntry("job-b", false))

	all := rl.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all[0].JobName != "job-a" {
		t.Errorf("expected first entry job-a, got %s", all[0].JobName)
	}
}

func TestRunLog_EvictsOldestWhenFull(t *testing.T) {
	rl := NewRunLog(3)
	rl.Append(makeEntry("first", true))
	rl.Append(makeEntry("second", true))
	rl.Append(makeEntry("third", true))
	rl.Append(makeEntry("fourth", true))

	all := rl.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(all))
	}
	if all[0].JobName != "second" {
		t.Errorf("expected oldest evicted, first entry should be 'second', got %s", all[0].JobName)
	}
}

func TestRunLog_ForJob_FiltersCorrectly(t *testing.T) {
	rl := NewRunLog(20)
	rl.Append(makeEntry("alpha", true))
	rl.Append(makeEntry("beta", false))
	rl.Append(makeEntry("alpha", false))

	alpha := rl.ForJob("alpha")
	if len(alpha) != 2 {
		t.Errorf("expected 2 entries for alpha, got %d", len(alpha))
	}
	beta := rl.ForJob("beta")
	if len(beta) != 1 {
		t.Errorf("expected 1 entry for beta, got %d", len(beta))
	}
	missing := rl.ForJob("gamma")
	if len(missing) != 0 {
		t.Errorf("expected 0 entries for unknown job, got %d", len(missing))
	}
}

func TestRunLog_Len(t *testing.T) {
	rl := NewRunLog(10)
	if rl.Len() != 0 {
		t.Errorf("expected 0 initial length")
	}
	rl.Append(makeEntry("x", true))
	rl.Append(makeEntry("y", true))
	if rl.Len() != 2 {
		t.Errorf("expected length 2, got %d", rl.Len())
	}
}
