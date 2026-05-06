package monitor

import (
	"testing"
	"time"
)

func TestNewDependencyStore_Empty(t *testing.T) {
	ds := NewDependencyStore()
	if len(ds.Edges()) != 0 {
		t.Fatal("expected no edges on init")
	}
}

func TestDependencyStore_BlockedWithNoEdges(t *testing.T) {
	ds := NewDependencyStore()
	if err := ds.Blocked("job-b", time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("expected no block, got: %v", err)
	}
}

func TestDependencyStore_BlockedWhenUpstreamNeverRan(t *testing.T) {
	ds := NewDependencyStore()
	ds.AddEdge("job-a", "job-b")
	err := ds.Blocked("job-b", time.Now().Add(-time.Hour))
	if err == nil {
		t.Fatal("expected block because upstream never ran")
	}
}

func TestDependencyStore_NotBlockedAfterUpstreamSucceeds(t *testing.T) {
	ds := NewDependencyStore()
	ds.AddEdge("job-a", "job-b")
	now := time.Now()
	ds.MarkSuccess("job-a", now)
	if err := ds.Blocked("job-b", now.Add(-time.Minute)); err != nil {
		t.Fatalf("expected no block, got: %v", err)
	}
}

func TestDependencyStore_BlockedWhenUpstreamTooOld(t *testing.T) {
	ds := NewDependencyStore()
	ds.AddEdge("job-a", "job-b")
	old := time.Now().Add(-2 * time.Hour)
	ds.MarkSuccess("job-a", old)
	// require success within the last 30 minutes
	err := ds.Blocked("job-b", time.Now().Add(-30*time.Minute))
	if err == nil {
		t.Fatal("expected block because upstream run is too old")
	}
}

func TestDependencyStore_MultipleUpstreams(t *testing.T) {
	ds := NewDependencyStore()
	ds.AddEdge("job-a", "job-c")
	ds.AddEdge("job-b", "job-c")
	now := time.Now()
	ds.MarkSuccess("job-a", now)
	// job-b hasn't run — job-c must still be blocked
	if err := ds.Blocked("job-c", now.Add(-time.Minute)); err == nil {
		t.Fatal("expected block because job-b has not succeeded")
	}
	ds.MarkSuccess("job-b", now)
	if err := ds.Blocked("job-c", now.Add(-time.Minute)); err != nil {
		t.Fatalf("expected no block after both upstreams succeed, got: %v", err)
	}
}

func TestDependencyStore_EdgesSnapshot(t *testing.T) {
	ds := NewDependencyStore()
	ds.AddEdge("a", "b")
	ds.AddEdge("b", "c")
	edges := ds.Edges()
	if len(edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(edges))
	}
}
