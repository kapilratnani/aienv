package log

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

var initOnce sync.Once

func init() { Init() }

func Init() {
	initOnce.Do(func() {
		home, err := os.UserHomeDir()
		if err != nil {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
			return
		}
		dir := filepath.Join(home, ".local", "aienv")
		if err := os.MkdirAll(dir, 0755); err != nil {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
			return
		}
		p := filepath.Join(dir, "aienv.log")
		f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
			return
		}
		w := io.Writer(f)
		slog.SetDefault(slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug})))
	})
}
