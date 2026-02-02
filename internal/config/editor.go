package config

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// EditorType defines the supported editor modes
type EditorType int

const (
	EditorSystem EditorType = iota
	EditorTerminal
)

// OpenEditor opens the configuration file in the specified editor
func OpenEditor(path string, editorType EditorType) error {
	var cmd *exec.Cmd

	switch editorType {
	case EditorSystem:
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", "-t", path)
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", "notepad", path)
		case "linux":
			cmd = exec.Command("xdg-open", path)
		default:
			return fmt.Errorf("unsupported platform for system editor: %s", runtime.GOOS)
		}
	case EditorTerminal:
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}
		// Check if editor exists
		if _, err := exec.LookPath(editor); err != nil {
			// Fallback to nano if vim not found
			editor = "nano"
		}
		
		cmd = exec.Command(editor, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if cmd == nil {
		return fmt.Errorf("failed to determine editor command")
	}

	return cmd.Run()
}
