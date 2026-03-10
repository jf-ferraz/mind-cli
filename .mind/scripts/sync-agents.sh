#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════
# sync-agents.sh — Sync conversation agents to Copilot platform
#
# Direction: .mind/conversation/agents/ → .github/agents/
#
# Primary source:
#   .mind/conversation/agents/ (Claude Code frontmatter + instruction body)
#
# Derived target:
#   .github/agents/ (Copilot Chat frontmatter + same instruction body)
#
# The script preserves Copilot-specific frontmatter in .github/agents/ and
# replaces only the instruction body with the content from .mind/.
#
# Usage:
#   bash .mind/scripts/sync-agents.sh          # sync all
#   bash .mind/scripts/sync-agents.sh --check  # dry-run, report diffs only
#
# Trigger: Run after any conversation agent body update in .mind/.
# ═══════════════════════════════════════════════════════════════
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SOURCE_DIR="$ROOT/.mind/conversation/agents"
TARGET_DIR="$ROOT/.github/agents"
CHECK_ONLY=false
SYNCED=0
DIFFS=0
ERRORS=0

if [[ "${1:-}" == "--check" ]]; then
    CHECK_ONLY=true
fi

# Filename mapping: .mind/ names → .github/ names
# .mind/conversation/agents/moderator.md → .github/agents/conversation-moderator.md
# .mind/conversation/agents/persona.md   → .github/agents/conversation-persona.md
# etc.
source_to_target_name() {
    local src_name="$1"
    echo "conversation-${src_name}"
}

# Extract instruction body (everything after the second ---)
extract_body() {
    local file="$1"
    awk 'BEGIN{c=0} /^---$/{c++; if(c==2){found=1; next}} found{print}' "$file"
}

# Extract frontmatter block (from first --- to second --- inclusive)
extract_frontmatter() {
    local file="$1"
    awk 'BEGIN{c=0} /^---$/{c++; if(c==2){print; exit}} {print}' "$file"
}

echo "=== Conversation Agent Sync ==="
echo "Source: $SOURCE_DIR (primary)"
echo "Target: $TARGET_DIR (Copilot)"
echo ""

for source_file in "$SOURCE_DIR"/*.md; do
    src_filename="$(basename "$source_file" .md)"
    target_filename="$(source_to_target_name "$src_filename").md"
    target_file="$TARGET_DIR/$target_filename"

    if [[ ! -f "$target_file" ]]; then
        echo "✗ $src_filename → $target_filename — target missing in .github/agents/"
        ERRORS=$((ERRORS + 1))
        continue
    fi

    source_body="$(extract_body "$source_file")"
    target_body="$(extract_body "$target_file")"

    if [[ "$source_body" == "$target_body" ]]; then
        echo "✓ $src_filename → $target_filename — in sync"
        SYNCED=$((SYNCED + 1))
    else
        DIFFS=$((DIFFS + 1))
        if $CHECK_ONLY; then
            echo "✗ $src_filename → $target_filename — body differs"
            diff <(echo "$source_body") <(echo "$target_body") | head -10
        else
            # Preserve Copilot frontmatter, replace body with .mind/ source
            target_frontmatter="$(extract_frontmatter "$target_file")"
            {
                echo "$target_frontmatter"
                echo "$source_body"
            } > "$target_file"
            echo "↻ $src_filename → $target_filename — synced"
            SYNCED=$((SYNCED + 1))
        fi
    fi
done

echo ""
echo "=== Summary ==="
echo "Synced: $SYNCED | Diffs: $DIFFS | Errors: $ERRORS"

if [[ $ERRORS -gt 0 ]]; then
    echo "⚠ $ERRORS target files missing in .github/agents/."
    exit 1
fi

if $CHECK_ONLY && [[ $DIFFS -gt 0 ]]; then
    echo "⚠ $DIFFS files have body drift. Run without --check to sync."
    exit 1
fi

exit 0
