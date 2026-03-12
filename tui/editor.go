package tui

import (
	"errors"
	"os"
	"os/exec"
)

// ErrEditorNotSet is returned when $EDITOR is not set.
var ErrEditorNotSet = errors.New("$EDITOR is not set — set the EDITOR environment variable to your preferred editor")

// editorCmd returns an exec.Cmd for opening a file in $EDITOR.
// Returns nil, ErrEditorNotSet if $EDITOR is unset.
func editorCmd(relPath string) (*exec.Cmd, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return nil, ErrEditorNotSet
	}
	return exec.Command(editor, relPath), nil
}
