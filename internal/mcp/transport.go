// Package mcp implements a Model Context Protocol server over stdio (JSON-RPC 2.0).
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Transport handles reading and writing JSON-RPC messages over stdio.
type Transport struct {
	reader *bufio.Reader
	writer io.Writer
}

// NewStdioTransport creates a transport that reads from stdin and writes to stdout.
func NewStdioTransport() *Transport {
	return &Transport{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// ReadMessage reads the next JSON-RPC message from the input stream.
// Returns io.EOF when the stream is closed.
func (t *Transport) ReadMessage() (json.RawMessage, error) {
	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		if err == io.EOF && len(line) == 0 {
			return nil, io.EOF
		}
		if err == io.EOF {
			return json.RawMessage(line), nil
		}
		return nil, fmt.Errorf("read message: %w", err)
	}
	return json.RawMessage(line), nil
}

// WriteMessage serializes v as JSON and writes it followed by a newline.
func (t *Transport) WriteMessage(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	data = append(data, '\n')
	_, err = t.writer.Write(data)
	return err
}
