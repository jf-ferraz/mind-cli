package generate

import (
	"fmt"
	"strings"
	"time"
)

// ADRTemplate returns content for a new ADR file.
func ADRTemplate(title string, seq int) string {
	return fmt.Sprintf(`# %d. %s

**Status**: Proposed
**Date**: %s

## Context

<!-- Describe the context and problem statement -->

## Decision

<!-- Describe the decision and rationale -->

## Consequences

<!-- Describe the consequences of this decision -->
`, seq, title, time.Now().Format("2006-01-02"))
}

// BlueprintTemplate returns content for a new blueprint file.
func BlueprintTemplate(title string, seq int) string {
	return fmt.Sprintf(`# BP-%02d: %s

## Overview

<!-- Describe the purpose and scope of this blueprint -->

## Design

<!-- Describe the design -->

## Implementation Notes

<!-- Implementation details and considerations -->
`, seq, title)
}

// IterationOverviewTemplate returns content for a new iteration overview.md.
func IterationOverviewTemplate(descriptor string, reqType string) string {
	title := strings.ReplaceAll(descriptor, "-", " ")
	return fmt.Sprintf(`# %s

- **Type**: %s
- **Request**: <!-- Describe the request -->
- **Agent Chain**: analyst → architect → developer → tester → reviewer
- **Branch**: <!-- branch name -->
- **Created**: %s

## Scope

<!-- Define the scope of this iteration -->

## Requirement Traceability

| Req ID | Description | Analyst | Architect | Developer | Reviewer |
|--------|-------------|---------|-----------|-----------|----------|
`, title, reqType, time.Now().Format("2006-01-02"))
}

// IterationChangesTemplate returns content for changes.md.
func IterationChangesTemplate() string {
	return `# Changes

| File | Change | Reason | Commit |
|------|--------|--------|--------|
`
}

// IterationTestSummaryTemplate returns content for test-summary.md.
func IterationTestSummaryTemplate() string {
	return `# Test Summary

## Test Results

| Suite | Pass | Fail | Skip | Total |
|-------|------|------|------|-------|

## Coverage

<!-- Coverage report -->
`
}

// IterationValidationTemplate returns content for validation.md.
func IterationValidationTemplate() string {
	return `# Validation

## Quality Gates

| Gate | Status | Notes |
|------|--------|-------|

## Checks

<!-- Validation check results -->
`
}

// IterationRetrospectiveTemplate returns content for retrospective.md.
func IterationRetrospectiveTemplate() string {
	return `# Retrospective

## What Went Well

<!-- List positives -->

## What Could Improve

<!-- List improvements -->

## Action Items

<!-- List actions -->
`
}

// SpikeTemplate returns content for a new spike report.
func SpikeTemplate(title string) string {
	return fmt.Sprintf(`# Spike: %s

**Date**: %s
**Time-box**: <!-- e.g., 2 hours -->

## Question

<!-- What question is this spike investigating? -->

## Findings

<!-- Summarize findings -->

## Recommendation

<!-- Recommended approach based on findings -->
`, title, time.Now().Format("2006-01-02"))
}

// ConvergenceTemplate returns content for a new convergence analysis.
func ConvergenceTemplate(title string) string {
	return fmt.Sprintf(`# Convergence: %s

**Date**: %s

## Context

<!-- Describe the analysis context -->

## Options Evaluated

<!-- List and evaluate options -->

## Decision Matrix

| Criterion | Weight | Option A | Option B | Option C |
|-----------|--------|----------|----------|----------|

## Recommendation

<!-- Final recommendation with rationale -->
`, title, time.Now().Format("2006-01-02"))
}

// BriefTemplate returns content for project-brief.md with filled sections.
func BriefTemplate(vision, deliverables, inScope, outScope, constraints string) string {
	return fmt.Sprintf(`# Project Brief

## Vision

%s

## Key Deliverables

%s

## Scope

### In Scope

%s

### Out of Scope

%s

## Constraints

%s
`, vision, deliverables, inScope, outScope, constraints)
}

// StubBriefTemplate returns a stub project-brief.md template.
func StubBriefTemplate() string {
	return `# Project Brief

## Vision

<!-- Describe the project vision -->

## Key Deliverables

<!-- List key deliverables -->

## Scope

### In Scope

<!-- What is in scope -->

### Out of Scope

<!-- What is out of scope -->

## Constraints

<!-- List constraints -->
`
}

// IndexEntry returns a markdown line for INDEX.md.
func IndexEntry(seq int, slug, filename string) string {
	title := strings.ReplaceAll(slug, "-", " ")
	return fmt.Sprintf("- [BP-%02d: %s](%s)\n", seq, title, filename)
}

// MindTomlTemplate returns the default mind.toml content.
// If frameworkVersion is non-empty, a [framework] section is included.
func MindTomlTemplate(name string, frameworkVersion string) string {
	frameworkSection := ""
	if frameworkVersion != "" {
		frameworkSection = fmt.Sprintf(`
[framework]
version = "%s"
mode = "standalone"
`, frameworkVersion)
	}

	return fmt.Sprintf(`[manifest]
schema = "mind/v1.0"
generation = 1
updated = %s

[project]
name = "%s"
description = ""
type = "cli"

[project.stack]
language = ""
framework = ""
testing = ""

[project.commands]
dev = ""
test = ""
lint = ""
typecheck = ""
build = ""
%s
[governance]
max-retries = 2
review-policy = ""
commit-policy = ""
branch-strategy = ""

[profiles]
active = []
`, time.Now().Format("2006-01-02T15:04:05Z"), name, frameworkSection)
}

// ClaudeAdapterTemplate returns the .claude/CLAUDE.md content.
func ClaudeAdapterTemplate() string {
	return `# Mind Agent Framework

> **All framework content lives in ` + "`.mind/`" + `.** This file is the Claude Code auto-loaded routing table.

Read ` + "`.mind/CLAUDE.md`" + ` for the full resource index with all agent paths, workflow types, and quality gates.
`
}

// StubDocument returns a generic stub document.
func StubDocument(title string) string {
	return fmt.Sprintf("# %s\n\n<!-- Content pending -->\n", title)
}

// WorkflowStub returns a stub workflow.md.
func WorkflowStub() string {
	return `# Workflow State

<!-- Workflow state is managed by the Mind Agent Framework -->
`
}

// CurrentStub returns a stub current.md.
func CurrentStub() string {
	return `# Current State

## Active Work

<!-- Describe current active work -->

## Recent Changes

<!-- List recent changes -->

## Known Issues

<!-- List known issues -->
`
}

// GlossaryStub returns a stub glossary.md.
func GlossaryStub() string {
	return `# Glossary

| Term | Definition |
|------|-----------|
`
}

// IndexStub returns a stub INDEX.md for blueprints.
func IndexStub() string {
	return `# Blueprints Index

<!-- Blueprint entries are auto-managed by mind create blueprint -->
`
}

// MCPConfigTemplate returns the .mcp.json content for Claude Code MCP discovery.
func MCPConfigTemplate() string {
	return `{
  "mcpServers": {
    "mind": {
      "command": "mind",
      "args": ["serve"],
      "env": {}
    }
  }
}
`
}
