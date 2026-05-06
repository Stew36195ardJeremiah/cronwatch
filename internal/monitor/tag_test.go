package monitor

import (
	"sort"
	"testing"
)

func TestNewTagStore_Empty(t *testing.T) {
	s := NewTagStore()
	if len(s.All()) != 0 {
		t.Fatal("expected empty store")
	}
}

func TestTagStore_SetAndGet(t *testing.T) {
	s := NewTagStore()
	s.Set("backup", []string{"critical", "nightly"})
	tags := s.Get("backup")
	sort.Strings(tags)
	if len(tags) != 2 || tags[0] != "critical" || tags[1] != "nightly" {
		t.Fatalf("unexpected tags: %v", tags)
	}
}

func TestTagStore_Add(t *testing.T) {
	s := NewTagStore()
	s.Add("sync", "infra")
	s.Add("sync", "daily")
	tags := s.Get("sync")
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
}

func TestTagStore_Add_NoDuplicates(t *testing.T) {
	s := NewTagStore()
	s.Add("job", "tag1")
	s.Add("job", "tag1")
	if len(s.Get("job")) != 1 {
		t.Fatal("expected deduplication")
	}
}

func TestTagStore_Remove(t *testing.T) {
	s := NewTagStore()
	s.Set("job", []string{"a", "b"})
	if err := s.Remove("job", "a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tags := s.Get("job")
	if len(tags) != 1 || tags[0] != "b" {
		t.Fatalf("unexpected tags after remove: %v", tags)
	}
}

func TestTagStore_Remove_UnknownJob(t *testing.T) {
	s := NewTagStore()
	if err := s.Remove("ghost", "tag"); err == nil {
		t.Fatal("expected error for unknown job")
	}
}

func TestTagStore_JobsWithTag(t *testing.T) {
	s := NewTagStore()
	s.Set("jobA", []string{"critical", "nightly"})
	s.Set("jobB", []string{"nightly"})
	s.Set("jobC", []string{"weekly"})
	jobs := s.JobsWithTag("nightly")
	sort.Strings(jobs)
	if len(jobs) != 2 || jobs[0] != "jobA" || jobs[1] != "jobB" {
		t.Fatalf("unexpected jobs: %v", jobs)
	}
}

func TestTagStore_All_Snapshot(t *testing.T) {
	s := NewTagStore()
	s.Set("j1", []string{"x"})
	s.Set("j2", []string{"y", "z"})
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(all))
	}
	// mutating snapshot should not affect store
	all["j1"] = append(all["j1"], "injected")
	if len(s.Get("j1")) != 1 {
		t.Fatal("snapshot mutation affected store")
	}
}
