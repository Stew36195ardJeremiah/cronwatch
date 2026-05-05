package monitor

import (
	"testing"
	"time"
)

func TestNewStatusStore_Empty(t *testing.T) {
	s := NewStatusStore()
	if len(s.All()) != 0 {
		t.Fatal("expected empty store")
	}
}

func TestStatusStore_SetAndGet(t *testing.T) {
	s := NewStatusStore()
	now := time.Now()
	s.Set("backup", JobStatus{
		LastRun: now,
		Healthy: true,
	})

	st, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected to find job status")
	}
	if st.Name != "backup" {
		t.Errorf("expected name 'backup', got %q", st.Name)
	}
	if !st.LastRun.Equal(now) {
		t.Errorf("expected LastRun %v, got %v", now, st.LastRun)
	}
	if !st.Healthy {
		t.Error("expected Healthy=true")
	}
}

func TestStatusStore_GetUnknown(t *testing.T) {
	s := NewStatusStore()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected not found for unknown job")
	}
}

func TestStatusStore_All_ReturnsSnapshot(t *testing.T) {
	s := NewStatusStore()
	s.Set("job1", JobStatus{Healthy: true})
	s.Set("job2", JobStatus{Healthy: false})

	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 statuses, got %d", len(all))
	}
}

func TestStatusStore_MarkFailed_NewJob(t *testing.T) {
	s := NewStatusStore()
	s.MarkFailed("newjob")

	st, ok := s.Get("newjob")
	if !ok {
		t.Fatal("expected job to be created on MarkFailed")
	}
	if st.FailCount != 1 {
		t.Errorf("expected FailCount=1, got %d", st.FailCount)
	}
	if st.Healthy {
		t.Error("expected Healthy=false after MarkFailed")
	}
}

func TestStatusStore_MarkFailed_Increments(t *testing.T) {
	s := NewStatusStore()
	s.Set("job", JobStatus{Healthy: true})
	s.MarkFailed("job")
	s.MarkFailed("job")

	st, _ := s.Get("job")
	if st.FailCount != 2 {
		t.Errorf("expected FailCount=2, got %d", st.FailCount)
	}
	if st.Healthy {
		t.Error("expected Healthy=false")
	}
}

func TestStatusStore_Reset_ClearsFailures(t *testing.T) {
	s := NewStatusStore()
	s.Set("job", JobStatus{FailCount: 3, Healthy: false})
	s.Reset("job")

	st, _ := s.Get("job")
	if st.FailCount != 0 {
		t.Errorf("expected FailCount=0, got %d", st.FailCount)
	}
	if !st.Healthy {
		t.Error("expected Healthy=true after Reset")
	}
}
