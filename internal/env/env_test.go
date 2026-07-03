package env

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestApplyPromptEmpty(t *testing.T) {
	e := &Env{Agent: AgentConfig{Args: []string{"--model", "claude"}}}
	e.ApplyPrompt("")
	if len(e.Agent.Args) != 2 {
		t.Errorf("expected no change, got %v", e.Agent.Args)
	}
}

func TestApplyPromptPositional(t *testing.T) {
	e := &Env{Agent: AgentConfig{Args: []string{"--model", "claude"}}}
	e.ApplyPrompt("Implement issue #123")
	want := []string{"--model", "claude", "Implement issue #123"}
	equal(t, e.Agent.Args, want)
}

func TestApplyPromptWithFlag(t *testing.T) {
	e := &Env{Agent: AgentConfig{PromptFlag: "-p", Args: []string{"--model", "claude"}}}
	e.ApplyPrompt("Implement issue #123")
	want := []string{"--model", "claude", "-p", "Implement issue #123"}
	equal(t, e.Agent.Args, want)
}

func TestApplyPromptWithFlagLong(t *testing.T) {
	e := &Env{Agent: AgentConfig{PromptFlag: "--message", Args: []string{}}}
	e.ApplyPrompt("fix bugs")
	want := []string{"--message", "fix bugs"}
	equal(t, e.Agent.Args, want)
}

func TestApplyPromptNoExistingArgs(t *testing.T) {
	e := &Env{Agent: AgentConfig{}}
	e.ApplyPrompt("hello")
	want := []string{"hello"}
	equal(t, e.Agent.Args, want)
}

func TestApplyPromptNoExistingArgsWithFlag(t *testing.T) {
	e := &Env{Agent: AgentConfig{PromptFlag: "-p"}}
	e.ApplyPrompt("hello")
	want := []string{"-p", "hello"}
	equal(t, e.Agent.Args, want)
}

func TestPromptFlagYAML(t *testing.T) {
	e := &Env{
		Meta:  EnvMeta{Name: "test"},
		Agent: AgentConfig{Command: []string{"opencode"}, PromptFlag: "-p"},
	}

	data, err := yaml.Marshal(e)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got Env
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.Agent.PromptFlag != "-p" {
		t.Errorf("PromptFlag = %q, want %q", got.Agent.PromptFlag, "-p")
	}
}

func TestPromptFlagEmptyOmitted(t *testing.T) {
	e := &Env{
		Meta:  EnvMeta{Name: "test"},
		Agent: AgentConfig{Command: []string{"opencode"}, Args: []string{"--model", "claude"}},
	}

	data, err := yaml.Marshal(e)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if containsString(string(data), "prompt_flag") {
		t.Error("expected prompt_flag to be omitted from YAML when empty")
	}
}

func TestPromptFlagLoadRoundTrip(t *testing.T) {
	yml := `env:
  name: test
agent:
  command: [claude]
  prompt_flag: "--message"
  mounts:
    - source: /tmp
      target: /workspace
`

	var e Env
	if err := yaml.Unmarshal([]byte(yml), &e); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if e.Agent.PromptFlag != "--message" {
		t.Errorf("PromptFlag = %q, want %q", e.Agent.PromptFlag, "--message")
	}
}

func equal(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v (len=%d), want %v (len=%d)", got, len(got), want, len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestExitSubcommandYAML(t *testing.T) {
	e := &Env{
		Meta:  EnvMeta{Name: "test"},
		Agent: AgentConfig{Command: []string{"opencode"}, ExitSubcommand: "run", PromptFlag: "-p"},
	}

	data, err := yaml.Marshal(e)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got Env
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.Agent.ExitSubcommand != "run" {
		t.Errorf("ExitSubcommand = %q, want %q", got.Agent.ExitSubcommand, "run")
	}
}

func TestExitSubcommandEmptyOmitted(t *testing.T) {
	e := &Env{
		Meta:  EnvMeta{Name: "test"},
		Agent: AgentConfig{Command: []string{"opencode"}},
	}

	data, err := yaml.Marshal(e)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if containsString(string(data), "exit_subcommand") {
		t.Error("expected exit_subcommand to be omitted from YAML when empty")
	}
}

func TestExitSubcommandLoad(t *testing.T) {
	yml := `env:
  name: test
agent:
  command: [opencode]
  prompt_flag: "-p"
  exit_subcommand: "run"
  mounts:
    - source: /tmp
      target: /workspace
`

	var e Env
	if err := yaml.Unmarshal([]byte(yml), &e); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if e.Agent.ExitSubcommand != "run" {
		t.Errorf("ExitSubcommand = %q, want %q", e.Agent.ExitSubcommand, "run")
	}
}

func containsString(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
