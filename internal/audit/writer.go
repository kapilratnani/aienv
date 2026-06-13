package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Writer struct {
	dir string
}

func NewWriter(dir string) *Writer {
	return &Writer{dir: dir}
}

func (w *Writer) WriteSessionMeta(meta SessionMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling session meta: %w", err)
	}
	return os.WriteFile(filepath.Join(w.dir, "session.meta.json"), data, 0644)
}

func (w *Writer) AppendNetworkEntry(entry NetworkEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshalling network entry: %w", err)
	}
	f, err := os.OpenFile(filepath.Join(w.dir, "network.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening network audit file: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("writing network entry: %w", err)
	}
	return nil
}

func (w *Writer) ListNetworkEntries() ([]NetworkEntry, error) {
	data, err := os.ReadFile(filepath.Join(w.dir, "network.jsonl"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading network audit file: %w", err)
	}
	var entries []NetworkEntry
	for _, line := range splitLines(data) {
		if len(line) == 0 {
			continue
		}
		var e NetworkEntry
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("parsing network entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
