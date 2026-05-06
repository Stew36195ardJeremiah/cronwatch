package monitor

import (
	"fmt"
	"sync"
)

// TagStore maps job names to a set of string tags for filtering and grouping.
type TagStore struct {
	mu   sync.RWMutex
	tags map[string]map[string]struct{}
}

// NewTagStore creates an empty TagStore.
func NewTagStore() *TagStore {
	return &TagStore{
		tags: make(map[string]map[string]struct{}),
	}
}

// Set replaces all tags for a job.
func (s *TagStore) Set(job string, tags []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	set := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		set[t] = struct{}{}
	}
	s.tags[job] = set
}

// Add appends a single tag to a job's tag set.
func (s *TagStore) Add(job, tag string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tags[job] == nil {
		s.tags[job] = make(map[string]struct{})
	}
	s.tags[job][tag] = struct{}{}
}

// Remove deletes a single tag from a job's tag set.
func (s *TagStore) Remove(job, tag string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	set, ok := s.tags[job]
	if !ok {
		return fmt.Errorf("job %q not found", job)
	}
	delete(set, tag)
	return nil
}

// Get returns a sorted slice of tags for a job.
func (s *TagStore) Get(job string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	set := s.tags[job]
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	return out
}

// JobsWithTag returns all job names that carry the given tag.
func (s *TagStore) JobsWithTag(tag string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var jobs []string
	for job, set := range s.tags {
		if _, ok := set[tag]; ok {
			jobs = append(jobs, job)
		}
	}
	return jobs
}

// All returns a snapshot of the full tag map (job -> []tag).
func (s *TagStore) All() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]string, len(s.tags))
	for job, set := range s.tags {
		tags := make([]string, 0, len(set))
		for t := range set {
			tags = append(tags, t)
		}
		out[job] = tags
	}
	return out
}
