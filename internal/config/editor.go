package config

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type EditorType int

const (
	EditorSystem EditorType = iota
	EditorTerminal
)

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
		if _, err := exec.LookPath(editor); err != nil {
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
