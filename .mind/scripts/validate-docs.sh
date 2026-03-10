#!/usr/bin/env bash
# ============================================================================
# validate-docs.sh — Documentation Structure Validator
# ============================================================================
# Validates the 5-zone docs/ structure for completeness and correctness.
#
# Usage:
#   validate-docs.sh                  # Validate current directory
#   validate-docs.sh /path/to/dir     # Validate target directory
#   validate-docs.sh --strict         # Fail on stubs too
#
# Exit code: 0 if no FAILs, 1 if any FAIL found.
# ============================================================================

set -euo pipefail

# --- Parse arguments ---
TARGET_DIR="."
STRICT=false

for arg in "$@"; do
    case "$arg" in
        --strict) STRICT=true ;;
        --help|-h)
            echo "Usage: validate-docs.sh [/path/to/dir] [--strict]"
            echo ""
            echo "  --strict    Treat stubs (files with only headings/comments) as failures"
            exit 0
            ;;
        *) TARGET_DIR="$arg" ;;
    esac
done

TARGET_DIR="$(cd "$TARGET_DIR" 2>/dev/null && pwd)"
DOCS_DIR="$TARGET_DIR/docs"

# --- Counters ---
PASS=0
FAIL=0
WARNINGS=0

check() {
    local name="$1"
    local result="$2"  # 0=pass, 1=fail
    local detail="${3:-}"

    if [[ "$result" -eq 0 ]]; then
        echo "  ✓ $name"
        PASS=$((PASS + 1))
    else
        echo "  ✗ $name"
        [[ -n "$detail" ]] && echo "    → $detail"
        FAIL=$((FAIL + 1))
    fi
}

warn() {
    local name="$1"
    local detail="${2:-}"
    echo "  ⚠ $name"
    [[ -n "$detail" ]] && echo "    → $detail"
    WARNINGS=$((WARNINGS + 1))
}

# Detect if a file is a stub (only headings, comments, empty lines, and table headers)
is_stub() {
    local file="$1"
    [ ! -f "$file" ] && return 1
    [ ! -s "$file" ] && return 0  # empty file is a stub

    # Count lines with real content: exclude blank lines, markdown headings,
    # HTML comments (<!-- ... -->), table separator rows, blockquotes, and
    # placeholder table rows (containing <!-- -->)
    local real_content
    real_content="$(grep -cEv '^[[:space:]]*$|^#+ |^<[!]--|^-->|^\|[-| :]+\|$|^> |^\|.*<[!]--.*-->.*\|' "$file" 2>/dev/null)" || real_content=0

    [ "$real_content" -le 2 ]
}

echo "=== Documentation Structure Validation ==="
echo "Target: $TARGET_DIR"
echo "Mode: $([ "$STRICT" = true ] && echo "strict" || echo "normal")"
echo ""

# ─── Check 1: docs/ directory exists ───
echo "[1/17] docs/ directory"
check "docs/ directory exists" "$([ -d "$DOCS_DIR" ] && echo 0 || echo 1)"

if [ ! -d "$DOCS_DIR" ]; then
    echo ""
    echo "=== Results ==="
    echo "Pass: $PASS | Fail: $FAIL | Warnings: $WARNINGS"
    echo ""
    echo "✗ docs/ directory missing — cannot continue."
    exit 1
fi

# ─── Check 2: All 5 zone directories exist ───
echo "[2/17] Zone directories"
missing_zones=""
for zone in spec blueprints state iterations knowledge; do
    if [ ! -d "$DOCS_DIR/$zone" ]; then
        missing_zones="$missing_zones $zone/"
    fi
done
check "All 5 zone directories exist" "$([ -z "$missing_zones" ] && echo 0 || echo 1)" "$missing_zones"

# ─── Check 3: Required spec files ───
echo "[3/17] Required spec files"
missing_spec=""
for f in project-brief.md requirements.md architecture.md; do
    if [ ! -f "$DOCS_DIR/spec/$f" ]; then
        missing_spec="$missing_spec $f"
    fi
done
check "Required spec files present" "$([ -z "$missing_spec" ] && echo 0 || echo 1)" "$missing_spec"

# ─── Check 4: decisions/ subdirectory ───
echo "[4/17] decisions/ subdirectory"
if [ -d "$DOCS_DIR/spec/decisions" ]; then
    check "decisions/ subdirectory exists" 0
else
    warn "decisions/ subdirectory missing" "Create with: mkdir -p docs/spec/decisions/"
fi

# ─── Check 5: ADR naming convention ───
echo "[5/17] ADR naming convention"
bad_adrs=""
if [ -d "$DOCS_DIR/spec/decisions" ]; then
    for f in "$DOCS_DIR/spec/decisions"/*.md; do
        [ -e "$f" ] || continue
        local_name="$(basename "$f")"
        # Skip _template.md
        [ "$local_name" = "_template.md" ] && continue
        # Must match NNN-descriptor.md (1+ leading digits)
        if ! echo "$local_name" | grep -qE '^[0-9]+-[a-z0-9].*\.md$'; then
            bad_adrs="$bad_adrs $local_name"
        fi
    done
fi
if [ -n "$bad_adrs" ]; then
    warn "ADR naming convention violated" "$bad_adrs (expected: NNN-descriptor.md)"
else
    check "ADR naming follows NNN-descriptor.md" 0
fi

# ─── Check 6: blueprints/INDEX.md ───
echo "[6/17] blueprints/INDEX.md"
check "blueprints/INDEX.md exists" "$([ -f "$DOCS_DIR/blueprints/INDEX.md" ] && echo 0 || echo 1)"

# ─── Check 7: Blueprint files have INDEX.md rows ───
echo "[7/17] Blueprint → INDEX.md coverage"
missing_index_rows=""
if [ -d "$DOCS_DIR/blueprints" ] && [ -f "$DOCS_DIR/blueprints/INDEX.md" ]; then
    for f in "$DOCS_DIR/blueprints"/[0-9]*.md; do
        [ -e "$f" ] || continue
        local_name="$(basename "$f")"
        if ! grep -q "$local_name" "$DOCS_DIR/blueprints/INDEX.md" 2>/dev/null; then
            missing_index_rows="$missing_index_rows $local_name"
        fi
    done
fi
if [ -n "$missing_index_rows" ]; then
    warn "Blueprint files without INDEX.md row" "$missing_index_rows"
else
    check "All blueprints have INDEX.md entries" 0
fi

# ─── Check 8: INDEX.md rows reference existing files ───
echo "[8/17] INDEX.md → file references"
bad_refs=""
if [ -f "$DOCS_DIR/blueprints/INDEX.md" ]; then
    # Extract file references from markdown links: [text](filename.md)
    while IFS= read -r ref; do
        if [ ! -f "$DOCS_DIR/blueprints/$ref" ]; then
            bad_refs="$bad_refs $ref"
        fi
    done < <(grep -oE '\([0-9][0-9a-z-]*\.md\)' "$DOCS_DIR/blueprints/INDEX.md" 2>/dev/null | tr -d '()' || true)
fi
check "INDEX.md references resolve" "$([ -z "$bad_refs" ] && echo 0 || echo 1)" "$bad_refs"

# ─── Check 9: state/current.md ───
echo "[9/17] state/current.md"
check "state/current.md exists" "$([ -f "$DOCS_DIR/state/current.md" ] && echo 0 || echo 1)"

# ─── Check 10: state/workflow.md ───
echo "[10/17] state/workflow.md"
if [ -f "$DOCS_DIR/state/workflow.md" ]; then
    check "state/workflow.md exists" 0
else
    warn "state/workflow.md missing" "Create with: docs-gen.sh or scaffold.sh"
fi

# ─── Check 11: knowledge/glossary.md ───
echo "[11/17] knowledge/glossary.md"
if [ -f "$DOCS_DIR/knowledge/glossary.md" ]; then
    check "knowledge/glossary.md exists" 0
else
    warn "knowledge/glossary.md missing" "Create with: scaffold.sh"
fi

# ─── Check 12: Iteration folder naming ───
echo "[12/17] Iteration folder naming"
bad_iters=""
if [ -d "$DOCS_DIR/iterations" ]; then
    for d in "$DOCS_DIR/iterations"/*/; do
        [ -d "$d" ] || continue
        local_name="$(basename "$d")"
        # Skip .gitkeep entries
        [ "$local_name" = "*" ] && continue
        # Must match NNN-TYPE-descriptor/ (type is uppercase)
        if ! echo "$local_name" | grep -qE '^[0-9]+-[A-Z_]+-[a-z0-9]'; then
            bad_iters="$bad_iters $local_name"
        fi
    done
fi
if [ -n "$bad_iters" ]; then
    warn "Iteration naming convention violated" "$bad_iters (expected: NNN-TYPE-descriptor/)"
else
    check "Iteration folders follow naming convention" 0
fi

# ─── Check 13: Each iteration has overview.md ───
echo "[13/17] Iteration overview.md presence"
missing_overview=""
if [ -d "$DOCS_DIR/iterations" ]; then
    for d in "$DOCS_DIR/iterations"/[0-9]*/; do
        [ -d "$d" ] || continue
        if [ ! -f "$d/overview.md" ]; then
            missing_overview="$missing_overview $(basename "$d")"
        fi
    done
fi
if [ -n "$missing_overview" ]; then
    warn "Iterations missing overview.md" "$missing_overview"
else
    check "All iterations have overview.md" 0
fi

# ─── Check 14: Spike files naming ───
echo "[14/17] Spike file naming"
bad_spikes=""
if [ -d "$DOCS_DIR/knowledge" ]; then
    for f in "$DOCS_DIR/knowledge"/*spike*; do
        [ -e "$f" ] || continue
        local_name="$(basename "$f")"
        if ! echo "$local_name" | grep -qE '\-spike\.md$'; then
            bad_spikes="$bad_spikes $local_name"
        fi
    done
fi
if [ -n "$bad_spikes" ]; then
    warn "Spike files not using -spike.md suffix" "$bad_spikes"
else
    check "Spike files use -spike.md suffix" 0
fi

# ─── Check 15: No legacy paths ───
echo "[15/17] No legacy paths"
legacy_paths=""
for p in adr adrs spikes architecture current; do
    if [ -d "$DOCS_DIR/$p" ]; then
        legacy_paths="$legacy_paths docs/$p/"
    fi
done
check "No legacy paths" "$([ -z "$legacy_paths" ] && echo 0 || echo 1)" "$legacy_paths"

# ─── Check 16: Stub detection ───
echo "[16/17] Stub detection"
stubs=""
# Check key files for stub content
for f in "$DOCS_DIR/spec/project-brief.md" "$DOCS_DIR/spec/requirements.md" "$DOCS_DIR/spec/architecture.md" "$DOCS_DIR/state/current.md"; do
    if [ -f "$f" ] && is_stub "$f"; then
        stubs="$stubs $(echo "$f" | sed "s|$TARGET_DIR/||")"
    fi
done
if [ -n "$stubs" ]; then
    if [ "$STRICT" = true ]; then
        check "No stub files" 1 "$stubs"
    else
        warn "Stub files detected (use --strict to fail on these)" "$stubs"
    fi
else
    check "No stub files in key documents" 0
fi

# ─── Check 17: Project brief minimum viable content ───
echo "[17/17] Project brief content completeness"
if [ -f "$DOCS_DIR/spec/project-brief.md" ] && ! is_stub "$DOCS_DIR/spec/project-brief.md"; then
    brief_missing=""
    for section in "Vision" "Key Deliverables" "Scope"; do
        if ! grep -qi "^##.*$section" "$DOCS_DIR/spec/project-brief.md" 2>/dev/null; then
            brief_missing="$brief_missing \"$section\""
        fi
    done
    if [ -n "$brief_missing" ]; then
        if [ "$STRICT" = true ]; then
            check "Project brief minimum sections" 1 "Missing:$brief_missing"
        else
            warn "Project brief missing key sections" "$brief_missing"
        fi
    else
        check "Project brief has minimum viable sections" 0
    fi
else
    check "Project brief content completeness" 0  # caught by stub/existence checks
fi

# ─── Summary ───
echo ""
echo "=== Results ==="
echo "Pass: $PASS | Fail: $FAIL | Warnings: $WARNINGS"
echo ""

if [[ $FAIL -gt 0 ]]; then
    echo "✗ $FAIL check(s) failed."
    exit 1
else
    echo "✓ All checks passed."
    exit 0
fi
