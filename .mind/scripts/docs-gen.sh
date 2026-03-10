#!/usr/bin/env bash
# ============================================================================
# docs-gen.sh — Incremental Document Generator
# ============================================================================
# Generates individual documents on demand with correct sequencing,
# naming, and template substitution.
#
# Usage:
#   docs-gen.sh adr "Chose PostgreSQL"
#   docs-gen.sh blueprint "API Gateway Design"
#   docs-gen.sh iteration enhancement "dashboard"
#   docs-gen.sh spike "Redis vs Memcached"
#   docs-gen.sh convergence "Auth Strategy"
#   docs-gen.sh list
#
# All output goes under docs/ in the current directory (or $TARGET_DIR).
# ============================================================================

set -euo pipefail

# --- Locate framework templates ---
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEMPLATE_DIR="$SCRIPT_DIR/../docs/templates"

# --- Target directory (default: current working directory) ---
TARGET_DIR="${DOCS_GEN_TARGET:-$(pwd)}"
DOCS_DIR="$TARGET_DIR/docs"
TODAY="$(date +%Y-%m-%d)"

# ============================================================================
# Utility functions
# ============================================================================

slugify() {
    echo "$1" | tr '[:upper:]' '[:lower:]' | \
        sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//'
}

# next_seq DIR PATTERN WIDTH
# Scans DIR for files/dirs matching PATTERN, extracts leading numbers,
# finds the max, and returns the next number zero-padded to WIDTH.
next_seq() {
    local dir="$1"
    local pattern="$2"  # glob pattern for ls
    local width="$3"

    local max=0
    if [ -d "$dir" ]; then
        for entry in "$dir"/$pattern; do
            [ -e "$entry" ] || continue
            local base
            base="$(basename "$entry")"
            # Extract leading digits (handles 001, 02, 0003, etc.)
            local num
            num="$(echo "$base" | sed -E 's/^0*([0-9]+).*/\1/')"
            if [ -n "$num" ] && [ "$num" -gt "$max" ] 2>/dev/null; then
                max="$num"
            fi
        done
    fi

    printf "%0${width}d" $((max + 1))
}

apply_template() {
    local template="$1"
    local output="$2"
    local title="$3"
    local seq="${4:-}"
    local type="${5:-}"

    if [ ! -f "$template" ]; then
        echo "Error: Template not found: $template" >&2
        exit 1
    fi

    # Escape sed special characters in title to prevent injection
    local escaped_title
    escaped_title="$(printf '%s' "$title" | sed -e 's/[\/&]/\\&/g')"

    sed \
        -e "s/__TITLE__/$escaped_title/g" \
        -e "s/__DATE__/$TODAY/g" \
        -e "s/__SEQ__/$seq/g" \
        -e "s/__TYPE__/$type/g" \
        "$template" > "$output"
}

# ============================================================================
# Subcommands
# ============================================================================

cmd_adr() {
    local title="$1"
    local slug
    slug="$(slugify "$title")"

    # Check if a file with this slug already exists (any sequence number)
    local existing=""
    for f in "$DOCS_DIR/spec/decisions/"*"-${slug}.md"; do
        [ -e "$f" ] && existing="$f" && break
    done
    if [ -n "$existing" ]; then
        echo "Exists: $existing (skipped)"
        return 0
    fi

    local seq
    seq="$(next_seq "$DOCS_DIR/spec/decisions" "[0-9]*.md" 3)"
    local output="$DOCS_DIR/spec/decisions/${seq}-${slug}.md"

    mkdir -p "$DOCS_DIR/spec/decisions"
    apply_template "$TEMPLATE_DIR/adr.md" "$output" "$title" "$seq"
    echo "Created: $output"
}

cmd_blueprint() {
    local title="$1"
    local slug
    slug="$(slugify "$title")"

    # Check if a file with this slug already exists (any sequence number)
    local existing=""
    for f in "$DOCS_DIR/blueprints/"*"-${slug}.md"; do
        [ -e "$f" ] && existing="$f" && break
    done
    if [ -n "$existing" ]; then
        echo "Exists: $existing (skipped)"
        return 0
    fi

    local seq
    seq="$(next_seq "$DOCS_DIR/blueprints" "[0-9]*.md" 2)"
    local output="$DOCS_DIR/blueprints/${seq}-${slug}.md"

    mkdir -p "$DOCS_DIR/blueprints"
    apply_template "$TEMPLATE_DIR/blueprint.md" "$output" "$title" "$seq"

    # Auto-update INDEX.md with new row (create if missing)
    local index="$DOCS_DIR/blueprints/INDEX.md"
    if [ ! -f "$index" ]; then
        cat > "$index" <<'INDEXEOF'
# Blueprints Index

## Active Blueprints

| # | Blueprint | Status | Summary |
|---|-----------|--------|---------|
INDEXEOF
        echo "Created: $index"
    fi
    if [ -f "$index" ]; then
        local row="| $seq | [${title}](${seq}-${slug}.md) | Active | <!-- summary --> |"
        # Append row after the last table row (before any blank line or ## heading after the table)
        # Find the last line matching the table pattern and append after it
        if grep -q "^|" "$index"; then
            # Append after the last | line in the Active section
            local last_table_line
            last_table_line="$(grep -n "^|" "$index" | tail -1 | cut -d: -f1)"
            sed -i "${last_table_line}a\\${row}" "$index"
            echo "Updated: $index (added row for $seq)"
        fi
    fi

    echo "Created: $output"
}

cmd_iteration() {
    local type_alias="$1"
    local descriptor="$2"
    local slug
    slug="$(slugify "$descriptor")"

    # Resolve type alias
    local type
    case "$type_alias" in
        new|new_project|NEW_PROJECT)        type="NEW_PROJECT" ;;
        enhancement|feature|ENHANCEMENT)    type="ENHANCEMENT" ;;
        bugfix|fix|bug|BUG_FIX)             type="BUG_FIX" ;;
        refactor|REFACTOR)                  type="REFACTOR" ;;
        *)
            echo "Error: Unknown iteration type '$type_alias'" >&2
            echo "  Valid types: new, enhancement/feature, bugfix/fix, refactor" >&2
            exit 1
            ;;
    esac

    # Check if an iteration with this slug already exists (any sequence/type)
    local existing=""
    for d in "$DOCS_DIR/iterations/"*"-${slug}"; do
        [ -d "$d" ] && existing="$d" && break
    done
    if [ -n "$existing" ]; then
        echo "Exists: $existing/ (skipped)"
        return 0
    fi

    local seq
    seq="$(next_seq "$DOCS_DIR/iterations" "[0-9]*" 3)"
    local dir_name="${seq}-${type}-${slug}"
    local iter_dir="$DOCS_DIR/iterations/$dir_name"

    mkdir -p "$iter_dir"

    local display_title
    display_title="$(echo "$descriptor" | sed 's/\b\(.\)/\u\1/g')"

    apply_template "$TEMPLATE_DIR/iteration-overview.md"       "$iter_dir/overview.md"       "$display_title" "$seq" "$type"
    apply_template "$TEMPLATE_DIR/iteration-changes.md"        "$iter_dir/changes.md"        "$display_title" "$seq" "$type"
    apply_template "$TEMPLATE_DIR/iteration-test-summary.md"   "$iter_dir/test-summary.md"   "$display_title" "$seq" "$type"
    apply_template "$TEMPLATE_DIR/iteration-validation.md"     "$iter_dir/validation.md"     "$display_title" "$seq" "$type"
    apply_template "$TEMPLATE_DIR/iteration-retrospective.md"  "$iter_dir/retrospective.md"  "$display_title" "$seq" "$type"

    echo "Created: $iter_dir/"
    echo "  overview.md, changes.md, test-summary.md, validation.md, retrospective.md"
}

cmd_spike() {
    local title="$1"
    local slug
    slug="$(slugify "$title")"

    local output="$DOCS_DIR/knowledge/${slug}-spike.md"

    if [ -f "$output" ]; then
        echo "Exists: $output (skipped)"
        return 0
    fi

    mkdir -p "$DOCS_DIR/knowledge"
    apply_template "$TEMPLATE_DIR/spike.md" "$output" "$title"
    echo "Created: $output"
}

cmd_convergence() {
    local title="$1"
    local slug
    slug="$(slugify "$title")"

    local output="$DOCS_DIR/knowledge/${slug}-convergence.md"

    if [ -f "$output" ]; then
        echo "Exists: $output (skipped)"
        return 0
    fi

    mkdir -p "$DOCS_DIR/knowledge"
    apply_template "$TEMPLATE_DIR/convergence.md" "$output" "$title"
    echo "Created: $output"
}

cmd_list() {
    echo "=== Documentation Audit ==="
    echo "Target: $DOCS_DIR"
    echo ""

    # Zone 1: spec/
    echo "[Zone 1: spec/]"
    for f in project-brief.md requirements.md architecture.md domain-model.md api-contracts.md; do
        if [ -f "$DOCS_DIR/spec/$f" ]; then
            echo "  present  $f"
        else
            echo "  MISSING  $f"
        fi
    done
    echo ""

    # ADRs
    echo "  decisions/"
    local adr_count=0
    if [ -d "$DOCS_DIR/spec/decisions" ]; then
        for f in "$DOCS_DIR/spec/decisions"/[0-9]*.md; do
            [ -e "$f" ] || continue
            echo "    $(basename "$f")"
            adr_count=$((adr_count + 1))
        done
    fi
    [ "$adr_count" -eq 0 ] && echo "    (none)"
    echo ""

    # Zone 2: blueprints/
    echo "[Zone 2: blueprints/]"
    if [ -f "$DOCS_DIR/blueprints/INDEX.md" ]; then
        echo "  present  INDEX.md"
    else
        echo "  MISSING  INDEX.md"
    fi
    local bp_count=0
    if [ -d "$DOCS_DIR/blueprints" ]; then
        for f in "$DOCS_DIR/blueprints"/[0-9]*.md; do
            [ -e "$f" ] || continue
            echo "  present  $(basename "$f")"
            bp_count=$((bp_count + 1))
        done
    fi
    echo ""

    # Zone 3: state/
    echo "[Zone 3: state/]"
    for f in current.md workflow.md; do
        if [ -f "$DOCS_DIR/state/$f" ]; then
            echo "  present  $f"
        else
            echo "  MISSING  $f"
        fi
    done
    echo ""

    # Zone 4: iterations/
    echo "[Zone 4: iterations/]"
    local iter_count=0
    if [ -d "$DOCS_DIR/iterations" ]; then
        for d in "$DOCS_DIR/iterations"/[0-9]*/; do
            [ -d "$d" ] || continue
            local base
            base="$(basename "$d")"
            local files=0
            for md in "$d"*.md; do
                [ -e "$md" ] && files=$((files + 1))
            done
            echo "  $base/ ($files files)"
            iter_count=$((iter_count + 1))
        done
    fi
    [ "$iter_count" -eq 0 ] && echo "  (none)"
    echo ""

    # Zone 5: knowledge/
    echo "[Zone 5: knowledge/]"
    if [ -f "$DOCS_DIR/knowledge/glossary.md" ]; then
        echo "  present  glossary.md"
    else
        echo "  MISSING  glossary.md"
    fi
    if [ -d "$DOCS_DIR/knowledge" ]; then
        for f in "$DOCS_DIR/knowledge"/*-spike.md; do
            [ -e "$f" ] || continue
            echo "  spike    $(basename "$f")"
        done
        for f in "$DOCS_DIR/knowledge"/*-convergence.md; do
            [ -e "$f" ] || continue
            echo "  convergence  $(basename "$f")"
        done
    fi
    echo ""

    # Legacy path warnings
    local legacy=0
    for p in "$DOCS_DIR/adr" "$DOCS_DIR/adrs" "$DOCS_DIR/spikes" "$DOCS_DIR/architecture" "$DOCS_DIR/current"; do
        if [ -d "$p" ]; then
            echo "WARNING: Legacy path detected: $(basename "$p")/"
            legacy=$((legacy + 1))
        fi
    done

    echo "--- Summary ---"
    echo "ADRs: $adr_count | Blueprints: $bp_count | Iterations: $iter_count"
    [ "$legacy" -gt 0 ] && echo "Legacy paths: $legacy (migrate these)"
}

# ============================================================================
# Main dispatch
# ============================================================================

if [ $# -eq 0 ]; then
    echo "Usage: docs-gen.sh <command> [args...]"
    echo ""
    echo "Commands:"
    echo "  adr <title>                  Create an Architecture Decision Record"
    echo "  blueprint <title>            Create a blueprint with INDEX.md update"
    echo "  iteration <type> <title>     Create an iteration folder (5 files)"
    echo "  spike <title>                Create a spike report"
    echo "  convergence <title>          Create a convergence analysis template"
    echo "  list                         Audit existing documentation"
    echo ""
    echo "Iteration types: new, enhancement/feature, bugfix/fix, refactor"
    echo ""
    echo "Environment:"
    echo "  DOCS_GEN_TARGET=<dir>        Override target directory (default: cwd)"
    exit 1
fi

CMD="$1"
shift

case "$CMD" in
    adr)
        [ $# -lt 1 ] && { echo "Usage: docs-gen.sh adr <title>" >&2; exit 1; }
        cmd_adr "$1"
        ;;
    blueprint)
        [ $# -lt 1 ] && { echo "Usage: docs-gen.sh blueprint <title>" >&2; exit 1; }
        cmd_blueprint "$1"
        ;;
    iteration)
        [ $# -lt 2 ] && { echo "Usage: docs-gen.sh iteration <type> <descriptor>" >&2; exit 1; }
        cmd_iteration "$1" "$2"
        ;;
    spike)
        [ $# -lt 1 ] && { echo "Usage: docs-gen.sh spike <title>" >&2; exit 1; }
        cmd_spike "$1"
        ;;
    convergence)
        [ $# -lt 1 ] && { echo "Usage: docs-gen.sh convergence <title>" >&2; exit 1; }
        cmd_convergence "$1"
        ;;
    list)
        cmd_list
        ;;
    *)
        echo "Error: Unknown command '$CMD'" >&2
        echo "Run 'docs-gen.sh' without arguments for usage." >&2
        exit 1
        ;;
esac
