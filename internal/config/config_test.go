package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	content := `
log_level: info
state_file: /tmp/cronwatch.state
smtp:
  host: smtp.example.com
  port: 587
  from: alerts@example.com
jobs:
  - name: backup
    schedule: "0 2 * * *"
    max_duration: 30m
    drift_threshold: 5m
    alert_email: ops@example.com
`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "backup" {
		t.Errorf("expected job name 'backup', got %q", cfg.Jobs[0].Name)
	}
	if cfg.Jobs[0].MaxDuration != 30*time.Minute {
		t.Errorf("expected max_duration 30m, got %v", cfg.Jobs[0].MaxDuration)
	}
}

func TestLoad_NoJobs(t *testing.T) {
	content := `log_level: debug\njobs: []\n`
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty jobs list")
	}
}

func TestLoad_MissingSchedule(t *testing.T) {
	content := `
jobs:
  - name: myjob
`
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing schedule")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/cronwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
