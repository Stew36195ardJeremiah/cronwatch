package monitor

import (
	"sync"
	"time"
)

// Annotation holds a timestamped note attached to a job.
type Annotation struct {
	JobName   string    `json:"job_name"`
	Author    string    `json:"author"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// AnnotationStore persists per-job annotations in memory.
type AnnotationStore struct {
	mu      sync.RWMutex
	entries map[string][]Annotation
	maxPer  int
}

// NewAnnotationStore creates a store that retains up to maxPer annotations per job.
func NewAnnotationStore(maxPer int) *AnnotationStore {
	if maxPer <= 0 {
		maxPer = 50
	}
	return &AnnotationStore{
		entries: make(map[string][]Annotation),
		maxPer:  maxPer,
	}
}

// Add appends an annotation for the given job, evicting the oldest if at capacity.
func (s *AnnotationStore) Add(a Annotation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	list := s.entries[a.JobName]
	if len(list) >= s.maxPer {
		list = list[1:]
	}
	s.entries[a.JobName] = append(list, a)
}

// Get returns all annotations for a job (newest last).
func (s *AnnotationStore) Get(jobName string) []Annotation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := s.entries[jobName]
	out := make([]Annotation, len(list))
	copy(out, list)
	return out
}

// Delete removes all annotations for a job.
func (s *AnnotationStore) Delete(jobName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, jobName)
}

// All returns a snapshot of every annotation across all jobs.
func (s *AnnotationStore) All() []Annotation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Annotation
	for _, list := range s.entries {
		out = append(out, list...)
	}
	return out
}
