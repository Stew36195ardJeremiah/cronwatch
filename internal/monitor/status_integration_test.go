package monitor_test

import (
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/monitor"
)

// TestStatusStore_ConcurrentAccess verifies the store is safe for concurrent use.
func TestStatusStore_ConcurrentAccess(t *testing.T) {
	s := monitor.NewStatusStore()
	done := make(chan struct{})

	go func() {
		for i := 0; i < 50; i++ {
			s.Set("jobA", monitor.JobStatus{Healthy: true, LastRun: time.Now()})
			s.MarkFailed("jobA")
		}
		close(done)
	}()

	for i := 0; i < 50; i++ {
		_ = s.All()
		_, _ = s.Get("jobA")
	}

	<-done
}

// TestStatusStore_OverdueReflectedInStatus ensures Overdue flag is preserved correctly.
func TestStatusStore_OverdueReflectedInStatus(t *testing.T) {
	s := monitor.NewStatusStore()

	s.Set("nightly", monitor.JobStatus{
		LastRun: time.Now().Add(-25 * time.Hour),
		Overdue: true,
		Drift:   25 * time.Hour,
		Healthy: false,
	})

	st, ok := s.Get("nightly")
	if !ok {
		t.Fatal("expected status for 'nightly'")
	}
	if !st.Overdue {
		t.Error("expected Overdue=true")
	}
	if st.Drift != 25*time.Hour {
		t.Errorf("expected drift 25h, got %v", st.Drift)
	}
}

// TestStatusStore_MultipleResets ensures repeated resets are idempotent.
func TestStatusStore_MultipleResets(t *testing.T) {
	s := monitor.NewStatusStore()
	s.Set("job", monitor.JobStatus{FailCount: 5, Healthy: false})
	s.Reset("job")
	s.Reset("job")

	st, _ := s.Get("job")
	if st.FailCount != 0 || !st.Healthy {
		t.Errorf("expected clean state after multiple resets, got %+v", st)
	}
}
