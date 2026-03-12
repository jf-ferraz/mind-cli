package repo

import (
	"bufio"
	"bytes"
	"strings"
)

// IsStubContent checks if content is a stub (only template boilerplate).
// A stub is a file with only headings, HTML comments, empty lines,
// table separators, blockquotes, and placeholder rows.
func IsStubContent(content []byte) bool {
	realLines := 0
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if isBoilerplateLine(line) {
			continue
		}
		realLines++
	}
	return realLines <= 2
}

func isBoilerplateLine(line string) bool {
	if line == "" {
		return true
	}
	if strings.HasPrefix(line, "#") {
		return true
	}
	if strings.HasPrefix(line, "<!--") || strings.HasPrefix(line, "-->") {
		return true
	}
	if strings.HasPrefix(line, ">") {
		return true
	}
	if isTableSeparator(line) {
		return true
	}
	if isPlaceholderRow(line) {
		return true
	}
	return false
}

func isTableSeparator(line string) bool {
	if !strings.HasPrefix(line, "|") {
		return false
	}
	cleaned := strings.ReplaceAll(line, "|", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned == ""
}

func isPlaceholderRow(line string) bool {
	return strings.HasPrefix(line, "|") &&
		strings.Contains(line, "<!--") &&
		strings.Contains(line, "-->")
}
