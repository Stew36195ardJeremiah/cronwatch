package monitor

import (
	"testing"
	"time"
)

func TestNewMetricsStore_Empty(t *testing.T) {
	ms := NewMetricsStore()
	if len(ms.All()) != 0 {
		t.Fatal("expected empty store")
	}
}

func TestMetricsStore_RecordAndGet(t *testing.T) {
	ms := NewMetricsStore()
	ms.Record("backup", 5*time.Second, false)

	met, ok := ms.Get("backup")
	if !ok {
		t.Fatal("expected metrics for 'backup'")
	}
	if met.RunCount != 1 {
		t.Errorf("RunCount = %d, want 1", met.RunCount)
	}
	if met.FailCount != 0 {
		t.Errorf("FailCount = %d, want 0", met.FailCount)
	}
	if met.LastDrift != 5*time.Second {
		t.Errorf("LastDrift = %v, want 5s", met.LastDrift)
	}
}

func TestMetricsStore_FailCount(t *testing.T) {
	ms := NewMetricsStore()
	ms.Record("deploy", 2*time.Second, false)
	ms.Record("deploy", 3*time.Second, true)

	met, _ := ms.Get("deploy")
	if met.RunCount != 2 {
		t.Errorf("RunCount = %d, want 2", met.RunCount)
	}
	if met.FailCount != 1 {
		t.Errorf("FailCount = %d, want 1", met.FailCount)
	}
}

func TestMetricsStore_MaxDrift(t *testing.T) {
	ms := NewMetricsStore()
	ms.Record("sync", 10*time.Second, false)
	ms.Record("sync", 30*time.Second, false)
	ms.Record("sync", 5*time.Second, false)

	met, _ := ms.Get("sync")
	if met.MaxDrift != 30*time.Second {
		t.Errorf("MaxDrift = %v, want 30s", met.MaxDrift)
	}
}

func TestMetricsStore_NegativeDriftAbsolute(t *testing.T) {
	ms := NewMetricsStore()
	ms.Record("early", -8*time.Second, false)

	met, _ := ms.Get("early")
	if met.MaxDrift != 8*time.Second {
		t.Errorf("MaxDrift = %v, want 8s (absolute)", met.MaxDrift)
	}
	if met.LastDrift != -8*time.Second {
		t.Errorf("LastDrift = %v, want -8s", met.LastDrift)
	}
}

func TestMetricsStore_All_ReturnsSnapshot(t *testing.T) {
	ms := NewMetricsStore()
	ms.Record("jobA", time.Second, false)
	ms.Record("jobB", 2*time.Second, false)

	all := ms.All()
	if len(all) != 2 {
		t.Errorf("All() len = %d, want 2", len(all))
	}
}

func TestMetricsStore_GetUnknown(t *testing.T) {
	ms := NewMetricsStore()
	_, ok := ms.Get("nonexistent")
	if ok {
		t.Fatal("expected not found for unknown job")
	}
}
