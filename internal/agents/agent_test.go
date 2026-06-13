package agents

import (
	"testing"

	"github.com/kapilratnani/aienv/internal/env"
)

type mockAgent struct{}

func (m *mockAgent) Name() string                                              { return "mock" }
func (m *mockAgent) GenerateFiles(e *env.Env, cwd string) ([]AgentFile, error) { return nil, nil }
func (m *mockAgent) DockerConfig(envDir string, e *env.Env, sessionID string) (*DockerConfig, error) {
	return &DockerConfig{}, nil
}

func TestRegisterAndGet(t *testing.T) {
	m := &mockAgent{}
	Register(m)

	got, err := Get("mock")
	if err != nil {
		t.Fatalf("Get('mock') unexpected error: %v", err)
	}
	if got.Name() != "mock" {
		t.Errorf("Get('mock').Name() = %q, want %q", got.Name(), "mock")
	}
}

func TestGetUnknown(t *testing.T) {
	_, err := Get("nonexistent")
	if err == nil {
		t.Fatal("Get('nonexistent') expected error, got nil")
	}
}
