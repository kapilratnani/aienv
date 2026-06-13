package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriterSessionMeta(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(dir)

	meta := SessionMeta{
		ID:           "test-session-1",
		EnvName:      "test-env",
		AgentCommand: "opencode",
		StartedAt:    time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC),
	}

	if err := w.WriteSessionMeta(meta); err != nil {
		t.Fatalf("WriteSessionMeta() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "session.meta.json"))
	if err != nil {
		t.Fatalf("reading session.meta.json: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty session.meta.json")
	}
}

func TestWriterAppendNetworkEntry(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(dir)

	entries := []NetworkEntry{
		{Timestamp: time.Now(), Host: "api.github.com", Method: "GET", Allowed: true},
		{Timestamp: time.Now(), Host: "evil.com", Method: "POST", Allowed: false},
	}

	for _, e := range entries {
		if err := w.AppendNetworkEntry(e); err != nil {
			t.Fatalf("AppendNetworkEntry() error: %v", err)
		}
	}

	got, err := w.ListNetworkEntries()
	if err != nil {
		t.Fatalf("ListNetworkEntries() error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}

	if got[0].Host != "api.github.com" {
		t.Errorf("expected host api.github.com, got %s", got[0].Host)
	}
	if got[1].Host != "evil.com" {
		t.Errorf("expected host evil.com, got %s", got[1].Host)
	}
}

func TestWriterAppendAppend(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(dir)

	w.AppendNetworkEntry(NetworkEntry{Host: "host1"})
	w.AppendNetworkEntry(NetworkEntry{Host: "host2"})

	entries, _ := w.ListNetworkEntries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestWriterListNoFile(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(dir)

	entries, err := w.ListNetworkEntries()
	if err != nil {
		t.Fatalf("ListNetworkEntries() error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}
