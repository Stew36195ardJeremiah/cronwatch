package monitor_test

import (
	"sync"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/monitor"
)

func TestSuppressionStore_ConcurrentAccess(t *testing.T) {
	s := monitor.NewSuppressionStore()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Suppress("job-x", time.Minute)
			_ = s.IsSuppressed("job-x")
			_ = s.All()
		}()
	}
	wg.Wait()

	if !s.IsSuppressed("job-x") {
		t.Fatal("expected job-x to still be suppressed after concurrent writes")
	}
}

func TestSuppressionStore_LiftDuringSuppression(t *testing.T) {
	s := monitor.NewSuppressionStore()
	s.Suppress("job-y", 10*time.Minute)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Millisecond)
		s.Lift("job-y")
	}()

	wg.Wait()
	if s.IsSuppressed("job-y") {
		t.Fatal("expected suppression to be lifted")
	}
}
