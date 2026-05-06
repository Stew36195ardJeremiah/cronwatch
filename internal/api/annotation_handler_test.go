package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/cronwatch/internal/monitor"
)

func newTestAnnotationStore() *monitor.AnnotationStore {
	return monitor.NewAnnotationStore(20)
}

func TestHandleAddAnnotation_Success(t *testing.T) {
	h := newAnnotationHandler(newTestAnnotationStore())
	body := `{"job_name":"backup","author":"alice","note":"post-deploy"}`
	req := httptest.NewRequest(http.MethodPost, "/annotations", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.handleAddAnnotation(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestHandleAddAnnotation_MissingFields(t *testing.T) {
	h := newAnnotationHandler(newTestAnnotationStore())
	body := `{"author":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/annotations", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.handleAddAnnotation(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleAddAnnotation_WrongMethod(t *testing.T) {
	h := newAnnotationHandler(newTestAnnotationStore())
	req := httptest.NewRequest(http.MethodGet, "/annotations", nil)
	w := httptest.NewRecorder()
	h.handleAddAnnotation(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleGetAnnotations_ReturnsEntries(t *testing.T) {
	store := newTestAnnotationStore()
	store.Add(monitor.Annotation{JobName: "backup", Author: "ops", Note: "note1"})
	h := newAnnotationHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/annotations?job=backup", nil)
	w := httptest.NewRecorder()
	h.handleGetAnnotations(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out []monitor.Annotation
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 1 || out[0].Note != "note1" {
		t.Errorf("unexpected annotations: %+v", out)
	}
}

func TestHandleGetAnnotations_MissingJobParam(t *testing.T) {
	h := newAnnotationHandler(newTestAnnotationStore())
	req := httptest.NewRequest(http.MethodGet, "/annotations", nil)
	w := httptest.NewRecorder()
	h.handleGetAnnotations(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleDeleteAnnotations_Success(t *testing.T) {
	store := newTestAnnotationStore()
	store.Add(monitor.Annotation{JobName: "cleanup", Author: "u", Note: "n"})
	h := newAnnotationHandler(store)
	req := httptest.NewRequest(http.MethodDelete, "/annotations?job=cleanup", nil)
	w := httptest.NewRecorder()
	h.handleDeleteAnnotations(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if len(store.Get("cleanup")) != 0 {
		t.Error("expected annotations deleted")
	}
}
