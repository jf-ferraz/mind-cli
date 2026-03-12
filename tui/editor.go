package tui

import (
	"os"
	"os/exec"
)

// editorCmd returns an exec.Cmd for opening a file in $EDITOR.
func editorCmd(relPath string) *exec.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	return exec.Command(editor, relPath)
}
