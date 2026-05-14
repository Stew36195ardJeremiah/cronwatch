package monitor

import (
	"testing"
)

func TestNewEnvStore_Empty(t *testing.T) {
	s := NewEnvStore()
	if len(s.All()) != 0 {
		t.Fatal("expected empty store")
	}
}

func TestEnvStore_SetAndGet(t *testing.T) {
	s := NewEnvStore()
	vars := map[string]string{"HOME": "/root", "PATH": "/usr/bin"}
	s.Set("backup", vars)

	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Job != "backup" {
		t.Errorf("expected job=backup, got %s", e.Job)
	}
	if e.Vars["HOME"] != "/root" {
		t.Errorf("expected HOME=/root, got %s", e.Vars["HOME"])
	}
	if e.CapturedAt.IsZero() {
		t.Error("expected CapturedAt to be set")
	}
}

func TestEnvStore_GetUnknownJob(t *testing.T) {
	s := NewEnvStore()
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestEnvStore_SetOverwrites(t *testing.T) {
	s := NewEnvStore()
	s.Set("sync", map[string]string{"X": "1"})
	s.Set("sync", map[string]string{"X": "2", "Y": "3"})

	e, _ := s.Get("sync")
	if e.Vars["X"] != "2" {
		t.Errorf("expected X=2, got %s", e.Vars["X"])
	}
	if e.Vars["Y"] != "3" {
		t.Errorf("expected Y=3, got %s", e.Vars["Y"])
	}
}

func TestEnvStore_Delete(t *testing.T) {
	s := NewEnvStore()
	s.Set("cleanup", map[string]string{"TMP": "/tmp"})
	s.Delete("cleanup")
	_, ok := s.Get("cleanup")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestEnvStore_All_ReturnsSnapshot(t *testing.T) {
	s := NewEnvStore()
	s.Set("job-a", map[string]string{"A": "1"})
	s.Set("job-b", map[string]string{"B": "2"})

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestEnvStore_IsolatesVarCopy(t *testing.T) {
	s := NewEnvStore()
	vars := map[string]string{"KEY": "original"}
	s.Set("isolated", vars)
	vars["KEY"] = "mutated"

	e, _ := s.Get("isolated")
	if e.Vars["KEY"] != "original" {
		t.Errorf("expected isolation, got %s", e.Vars["KEY"])
	}
}
