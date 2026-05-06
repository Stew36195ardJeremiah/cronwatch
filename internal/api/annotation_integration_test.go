package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

func TestAnnotation_EndToEnd_AddAndRetrieve(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}

	store := monitor.NewAnnotationStore(20)
	srv := newTestServer(t, port)

	ah := newAnnotationHandler(store)
	srv.Mux().Handle("/annotations", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			ah.handleAddAnnotation(w, r)
		case http.MethodGet:
			ah.handleGetAnnotations(w, r)
		case http.MethodDelete:
			ah.handleDeleteAnnotations(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	go srv.Start()
	time.Sleep(50 * time.Millisecond)
	defer srv.Stop()

	base := fmt.Sprintf("http://127.0.0.1:%d", port)

	// POST annotation
	body := `{"job_name":"nightly","author":"ci","note":"deployed v2"}`
	resp, err := http.Post(base+"/annotations", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	// GET annotations
	resp, err = http.Get(base + "/annotations?job=nightly")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	var list []monitor.Annotation
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) != 1 || list[0].Note != "deployed v2" {
		t.Errorf("unexpected list: %+v", list)
	}
}
