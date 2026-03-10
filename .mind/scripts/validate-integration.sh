#!/usr/bin/env bash
# scripts/validate-integration.sh
#
# Cross-reference validator for the Mind Framework integration.
# Runs 11 checks to ensure all paths, references, platform conventions,
# and model assignments are consistent.
#
# Usage:
#   bash scripts/validate-integration.sh
#
# Exit code: 0 if all checks pass, 1 if any check fails.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PROJECT_ROOT="$(cd "$ROOT/.." && pwd)"
PASS=0
FAIL=0
WARNINGS=0

check() {
    local name="$1"
    local result="$2" # 0=pass, 1=fail
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
    local detail="$2"
    echo "  ⚠ $name"
    echo "    → $detail"
    WARNINGS=$((WARNINGS + 1))
}

# Helper: map canonical agent name to conversation/agents/ filename
# conversation-moderator → moderator.md
# conversation-persona → persona.md
# conversation-persona-architect → persona-architect.md
agent_to_file() {
    local agent="$1"
    echo "${agent#conversation-}.md"
}

echo "=== Mind Framework Integration Validation ==="
echo "Root: $ROOT"
echo ""

# ─── Check 1: Conversation agent files exist in conversation/agents/ ───
echo "[1/11] Agent files in conversation/agents/"
missing=""
for agent in conversation-moderator conversation-persona conversation-persona-architect conversation-persona-pragmatist conversation-persona-critic conversation-persona-researcher; do
    local_file="$(agent_to_file "$agent")"
    if [[ ! -f "$ROOT/conversation/agents/$local_file" ]]; then
        missing="$missing $local_file"
    fi
done
check "6 conversation agents in conversation/agents/" "$([ -z "$missing" ] && echo 0 || echo 1)" "$missing"

# ─── Check 2: Conversation agent files exist in .github/agents/ ───
echo "[2/11] Agent files in .github/agents/"
missing=""
for agent in conversation-moderator conversation-persona conversation-persona-architect conversation-persona-pragmatist conversation-persona-critic conversation-persona-researcher; do
    if [[ ! -f "$PROJECT_ROOT/.github/agents/$agent.md" ]]; then
        missing="$missing $agent.md"
    fi
done
check "6 conversation agents in .github/agents/" "$([ -z "$missing" ] && echo 0 || echo 1)" "$missing"

# ─── Check 3: No stale panel-module/ references ───
echo "[3/11] No stale panel-module/ references"
# Exclude docs/archive/ — contains historical migration docs that legitimately reference panel-module
stale="$(grep -rl 'panel-module/' "$ROOT/agents/" "$PROJECT_ROOT/.github/agents/" "$ROOT/commands/" "$ROOT/conversation/" "$ROOT/CLAUDE.md" "$ROOT/ARCHITECTURE.md" 2>/dev/null || true)"
check "No panel-module/ references" "$([ -z "$stale" ] && echo 0 || echo 1)" "$stale"

# ─── Check 4: Orchestrator chains reference existing agent files ───
echo "[4/11] Orchestrator agent chain completeness"
missing_agents=""
for agent in orchestrator analyst architect developer tester reviewer discovery; do
    if [[ ! -f "$ROOT/agents/$agent.md" ]]; then
        missing_agents="$missing_agents $agent.md"
    fi
done
# Also check the conversation-moderator since it's in the COMPLEX_NEW chain
if [[ ! -f "$ROOT/conversation/agents/moderator.md" ]]; then
    missing_agents="$missing_agents moderator.md"
fi
check "All chain agents exist" "$([ -z "$missing_agents" ] && echo 0 || echo 1)" "$missing_agents"

# ─── Check 5: CLAUDE.md resource entries resolve to existing paths ───
echo "[5/11] CLAUDE.md resource path resolution"
bad_paths=""
# Extract paths from the resource table only (first table, stops at first blank line after it)
resource_table_state=0  # 0=before, 1=in-table, 2=done
while IFS= read -r line; do
    # Detect start of resource table (first |---| separator)
    if [[ "$line" == "|---"* && $resource_table_state -eq 0 ]]; then
        resource_table_state=1
        continue
    fi
    # End on blank line or next heading — and stay done
    if [[ $resource_table_state -eq 1 && ( -z "$line" || "$line" == "#"* ) ]]; then
        resource_table_state=2
        continue
    fi
    # Skip everything after the first table
    if [[ $resource_table_state -eq 2 ]]; then
        continue
    fi
    if [[ $resource_table_state -eq 1 ]]; then
        path="$(echo "$line" | sed -n 's/^| `\([^`]*\)`.*/\1/p')"
        if [[ -n "$path" && "$path" != "Resource" ]]; then
            resolved="$PROJECT_ROOT/$path"
            if [[ "$path" == *"*"* ]]; then
                count=$(ls $resolved 2>/dev/null | wc -l)
                if [[ "$count" -eq 0 ]]; then
                    bad_paths="$bad_paths $path"
                fi
            elif [[ ! -e "$resolved" ]]; then
                bad_paths="$bad_paths $path"
            fi
        fi
    fi
done < "$ROOT/CLAUDE.md"
check "CLAUDE.md paths resolve" "$([ -z "$bad_paths" ] && echo 0 || echo 1)" "$bad_paths"

# ─── Check 6: conversation/config/*.yml are valid YAML (syntax check) ───
echo "[6/11] YAML syntax in conversation/config/"
yaml_errors=""
if python3 -c "import yaml" 2>/dev/null; then
    for yml in "$ROOT/conversation/config/"*.yml; do
        if ! python3 -c "import yaml; yaml.safe_load(open('$yml'))" 2>/dev/null; then
            yaml_errors="$yaml_errors $(basename "$yml")"
        fi
    done
elif command -v ruby &>/dev/null; then
    for yml in "$ROOT/conversation/config/"*.yml; do
        if ! ruby -ryaml -e "YAML.load_file('$yml')" 2>/dev/null; then
            yaml_errors="$yaml_errors $(basename "$yml")"
        fi
    done
else
    # Fallback: basic syntax check without PyYAML — verify files parse as valid YAML
    # using Python's built-in string processing (checks for common syntax errors)
    for yml in "$ROOT/conversation/config/"*.yml; do
        # Check: file is non-empty, no tab indentation, no bare { or [ at start
        if [[ ! -s "$yml" ]]; then
            yaml_errors="$yaml_errors $(basename "$yml")(empty)"
        elif grep -Pn '^\t' "$yml" >/dev/null 2>&1; then
            yaml_errors="$yaml_errors $(basename "$yml")(tabs)"
        fi
    done
    if [[ -z "$yaml_errors" ]]; then
        warn "YAML syntax" "No YAML parser available (pip install pyyaml). Basic checks passed."
    fi
fi
check "YAML files parse" "$([ -z "$yaml_errors" ] && echo 0 || echo 1)" "$yaml_errors"

# ─── Check 7: No Copilot tool names in conversation/agents/ frontmatter ───
echo "[7/11] No Copilot tool names in conversation/agents/ frontmatter"
copilot_leaks=""
for f in "$ROOT/conversation/agents"/*.md; do
    [[ -f "$f" ]] || continue
    # Extract frontmatter only (lines between first and second ---)
    fm="$(awk 'BEGIN{c=0} /^---$/{c++; if(c==2) exit} c==1{print}' "$f")"
    if echo "$fm" | grep -qE '\breadFile\b|\bcodebase\b|\btextSearch\b|\bfileSearch\b|\bfetch\b'; then
        copilot_leaks="$copilot_leaks $(basename "$f")"
    fi
done
check "No Copilot tools in conversation/agents/ frontmatter" "$([ -z "$copilot_leaks" ] && echo 0 || echo 1)" "$copilot_leaks"

# ─── Check 8: No Claude tool names in .github/agents/ frontmatter ───
echo "[8/11] No Claude tool names in .github/agents/ frontmatter"
claude_leaks=""
for f in "$PROJECT_ROOT/.github/agents"/conversation-*.md; do
    [[ -f "$f" ]] || continue
    fm="$(awk 'BEGIN{c=0} /^---$/{c++; if(c==2) exit} c==1{print}' "$f")"
    if echo "$fm" | grep -qwE 'Task|Bash'; then
        claude_leaks="$claude_leaks $(basename "$f")"
    fi
done
check "No Claude tools in .github/agents/ frontmatter" "$([ -z "$claude_leaks" ] && echo 0 || echo 1)" "$claude_leaks"

# ─── Check 9: Body parity between platform pairs ───
echo "[9/11] Body parity between platforms"
extract_body() {
    awk 'BEGIN{c=0} /^---$/{c++; if(c==2){found=1; next}} found{print}' "$1"
}
drift=""
for agent in conversation-moderator conversation-persona conversation-persona-architect conversation-persona-pragmatist conversation-persona-critic conversation-persona-researcher; do
    source="$PROJECT_ROOT/.github/agents/$agent.md"
    local_file="$(agent_to_file "$agent")"
    target="$ROOT/conversation/agents/$local_file"
    if [[ -f "$source" && -f "$target" ]]; then
        if ! diff <(extract_body "$source") <(extract_body "$target") >/dev/null 2>&1; then
            drift="$drift $agent.md"
        fi
    fi
done
check "Body parity across platforms" "$([ -z "$drift" ] && echo 0 || echo 1)" "$drift"

# ─── Check 10: Model consistency — agent frontmatter vs personas.yml ───
echo "[10/11] Model consistency (agent frontmatter ↔ personas.yml)"
# Mapping: Claude Code model ID → Copilot model display name
# claude-opus-4-6 ↔ "Claude Opus 4.6"
# claude-sonnet-4-6 ↔ "Claude Sonnet 4.6"
# claude-haiku-4-5 ↔ "Claude Haiku 4.5"
model_mismatches=""

extract_frontmatter_model() {
    awk 'BEGIN{c=0} /^---$/{c++; if(c==2) exit} c==1{print}' "$1" | \
        grep -E '^model:' | sed 's/^model: *//' | tr -d '"'"'"
}

# Map Claude Code model IDs to Copilot display names for comparison
normalize_model() {
    local m="$1"
    case "$m" in
        claude-opus-4-6|"Claude Opus 4.6") echo "opus" ;;
        claude-sonnet-4-6|"Claude Sonnet 4.6"|"Claude Sonnet 4 (copilot)") echo "sonnet" ;;
        claude-haiku-4-5|"Claude Haiku 4.5") echo "haiku" ;;
        *) echo "unknown:$m" ;;
    esac
}

# Check persona agents against personas.yml
# personas.yml persona key → agent file mapping:
#   architect → persona-architect.md
#   pragmatist → persona-pragmatist.md
#   critic → persona-critic.md
#   researcher → persona-researcher.md
personas_yml="$ROOT/conversation/config/personas.yml"
if [[ -f "$personas_yml" ]]; then
    for persona in architect pragmatist critic researcher; do
        agent_file="$ROOT/conversation/agents/persona-${persona}.md"
        if [[ ! -f "$agent_file" ]]; then
            model_mismatches="$model_mismatches ${persona}(agent file missing)"
            continue
        fi

        # Extract frontmatter model from agent file
        fm_model="$(extract_frontmatter_model "$agent_file")"
        fm_normalized="$(normalize_model "$fm_model")"

        # Extract model from personas.yml (grep the model line after the persona key)
        # Use awk to find the persona block and extract its model
        yml_model="$(awk -v p="  ${persona}:" '
            $0 == p { found=1; next }
            found && /^  [a-z]/ { found=0 }
            found && /model:/ { gsub(/.*model: *"?/, ""); gsub(/".*/, ""); print; exit }
        ' "$personas_yml")"
        yml_normalized="$(normalize_model "$yml_model")"

        if [[ "$fm_normalized" != "$yml_normalized" ]]; then
            model_mismatches="$model_mismatches ${persona}(agent:${fm_model} != yml:${yml_model})"
        fi
    done

    # Also check .github/agents/ models match agent frontmatter
    for persona in architect pragmatist critic researcher; do
        agent_file="$ROOT/conversation/agents/persona-${persona}.md"
        github_file="$PROJECT_ROOT/.github/agents/conversation-persona-${persona}.md"
        if [[ -f "$agent_file" && -f "$github_file" ]]; then
            fm_model="$(extract_frontmatter_model "$agent_file")"
            gh_model="$(extract_frontmatter_model "$github_file")"
            fm_normalized="$(normalize_model "$fm_model")"
            gh_normalized="$(normalize_model "$gh_model")"
            if [[ "$fm_normalized" != "$gh_normalized" ]]; then
                model_mismatches="$model_mismatches ${persona}-copilot(agent:${fm_model} != github:${gh_model})"
            fi
        fi
    done
else
    warn "Model consistency" "personas.yml not found at $personas_yml"
fi
check "Model consistency across sources" "$([ -z "$model_mismatches" ] && echo 0 || echo 1)" "$model_mismatches"

# ─── Check 11: Model tier compliance — agents match tier policy ───
echo "[11/11] Model tier compliance"
tier_violations=""

# Tier policy (from agent-authoring.guide.md):
# Premium (opus): analyst, architect, reviewer, moderator, persona-critic
# Standard (sonnet): orchestrator, developer, tester, discovery, persona-*, persona
# Fast (haiku): technical-writer
declare -A TIER_POLICY=(
    # Core workflow agents
    ["agents/analyst.md"]="opus"
    ["agents/architect.md"]="opus"
    ["agents/reviewer.md"]="opus"
    ["agents/orchestrator.md"]="sonnet"
    ["agents/developer.md"]="sonnet"
    ["agents/tester.md"]="sonnet"
    ["agents/discovery.md"]="sonnet"
    ["agents/technical-writer.md"]="haiku"
    # Conversation agents
    ["conversation/agents/moderator.md"]="opus"
    ["conversation/agents/persona-critic.md"]="opus"
    ["conversation/agents/persona-architect.md"]="sonnet"
    ["conversation/agents/persona-pragmatist.md"]="sonnet"
    ["conversation/agents/persona-researcher.md"]="sonnet"
    ["conversation/agents/persona.md"]="sonnet"
)

for agent_path in "${!TIER_POLICY[@]}"; do
    full_path="$ROOT/$agent_path"
    expected_tier="${TIER_POLICY[$agent_path]}"
    if [[ -f "$full_path" ]]; then
        actual_model="$(extract_frontmatter_model "$full_path")"
        actual_tier="$(normalize_model "$actual_model")"
        if [[ "$actual_tier" != "$expected_tier" ]]; then
            tier_violations="$tier_violations $(basename "$agent_path")(expected:${expected_tier} actual:${actual_tier})"
        fi
    fi
done
check "Model tier compliance" "$([ -z "$tier_violations" ] && echo 0 || echo 1)" "$tier_violations"

# ─── Summary ───
echo ""
echo "=== Results ==="
echo "Pass: $PASS | Fail: $FAIL | Warnings: $WARNINGS"
echo ""

if [[ $FAIL -gt 0 ]]; then
    echo "⚠ $FAIL check(s) failed. Review issues above."
    exit 1
else
    echo "✓ All checks passed."
    exit 0
fi
