package monitor

import (
	"sync"
	"time"
)

// NoteEntry holds a freeform operator note attached to a job.
type NoteEntry struct {
	Job       string    `json:"job"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	Author    string    `json:"author,omitempty"`
}

// NoteStore stores the latest operator note per job.
type NoteStore struct {
	mu    sync.RWMutex
	notes map[string]NoteEntry
}

// NewNoteStore returns an initialised NoteStore.
func NewNoteStore() *NoteStore {
	return &NoteStore{
		notes: make(map[string]NoteEntry),
	}
}

// Set records or replaces the note for a job.
func (s *NoteStore) Set(job, note, author string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notes[job] = NoteEntry{
		Job:       job,
		Note:      note,
		Author:    author,
		CreatedAt: time.Now(),
	}
}

// Get returns the note for a job and whether one exists.
func (s *NoteStore) Get(job string) (NoteEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.notes[job]
	return e, ok
}

// Delete removes the note for a job.
func (s *NoteStore) Delete(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.notes, job)
}

// All returns a snapshot of all notes.
func (s *NoteStore) All() []NoteEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]NoteEntry, 0, len(s.notes))
	for _, e := range s.notes {
		out = append(out, e)
	}
	return out
}
