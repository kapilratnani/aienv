package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const shellFunc = `# aienv - AI environment activation
aienv() {
  case "$1" in
    create|list|show|init|edit|delete|activate|help|--help|-h)
      command aienv "$@"
      ;;
    *)
      eval "$(command aienv activate "$@")"
      ;;
  esac
}
`

func Detect() string {
	shell := os.Getenv("SHELL")
	switch {
	case shell == "":
		return "bash"
	case filepath.Base(shell) == "zsh":
		return "zsh"
	default:
		return "bash"
	}
}

func RCFile(shell string) string {
	home, _ := os.UserHomeDir()
	switch shell {
	case "zsh":
		return filepath.Join(home, ".zshrc")
	default:
		return filepath.Join(home, ".bashrc")
	}
}

func IsInstalled(rcPath string) (bool, error) {
	data, err := os.ReadFile(rcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return strings.Contains(string(data), "# aienv"), nil
}

func Install(rcPath string) error {
	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening %s: %w", rcPath, err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n" + shellFunc); err != nil {
		return fmt.Errorf("writing to %s: %w", rcPath, err)
	}
	return nil
}
