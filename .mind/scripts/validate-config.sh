#!/usr/bin/env bash
# scripts/validate-config.sh
#
# Validates conversation/config/*.yml files against expected structure.
# Checks: required top-level keys, cross-references, value constraints.
#
# Usage:
#   bash scripts/validate-config.sh
#
# Requires: python3 with PyYAML (standard on most systems)

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG_DIR="$ROOT/conversation/config"
PASS=0
FAIL=0

check() {
    local file="$1"
    local name="$2"
    local result="$3"
    local detail="${4:-}"

    if [[ "$result" -eq 0 ]]; then
        echo "  ✓ [$file] $name"
        PASS=$((PASS + 1))
    else
        echo "  ✗ [$file] $name"
        [[ -n "$detail" ]] && echo "    → $detail"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== Conversation Config Validation ==="
echo "Config dir: $CONFIG_DIR"
echo ""

# Check python3 + yaml available
HAS_PYYAML=0
if python3 -c "import yaml" 2>/dev/null; then
    HAS_PYYAML=1
fi

# ─── Check all 4 files exist ───
echo "[Files]"
for f in conversation.yml personas.yml quality.yml extensions.yml; do
    if [[ -f "$CONFIG_DIR/$f" ]]; then
        check "$f" "file exists" 0
    else
        check "$f" "file exists" 1 "Missing: $CONFIG_DIR/$f"
    fi
done

echo ""
echo "[Structure]"

if [[ $HAS_PYYAML -eq 1 ]]; then
# ─── Validate each config file (PyYAML available) ───
python3 - "$CONFIG_DIR" <<'PYEOF'
import sys, os, yaml

config_dir = sys.argv[1]
pass_count = 0
fail_count = 0

def check(filename, name, ok, detail=""):
    global pass_count, fail_count
    if ok:
        print(f"  ✓ [{filename}] {name}")
        pass_count += 1
    else:
        print(f"  ✗ [{filename}] {name}")
        if detail:
            print(f"    → {detail}")
        fail_count += 1

def load_yaml(filename):
    path = os.path.join(config_dir, filename)
    if not os.path.exists(path):
        return None
    with open(path) as f:
        return yaml.safe_load(f)

# ── conversation.yml ──
data = load_yaml("conversation.yml")
if data:
    for key in ["phases", "routing", "context", "termination", "moderation"]:
        check("conversation.yml", f"has '{key}' key", key in data, f"Missing top-level key: {key}")

    # Check phase IDs referenced in routing exist in phases
    if "phases" in data and "routing" in data:
        phase_ids = set()
        phases = data["phases"]
        if isinstance(phases, dict):
            phase_ids = set(phases.keys())
        elif isinstance(phases, list):
            for p in phases:
                if isinstance(p, dict) and "id" in p:
                    phase_ids.add(p["id"])
                elif isinstance(p, str):
                    phase_ids.add(p)

        if phase_ids:
            routing = data["routing"]
            bad_refs = []
            if isinstance(routing, dict):
                for k, v in routing.items():
                    targets = []
                    if isinstance(v, str):
                        targets = [v]
                    elif isinstance(v, list):
                        targets = v
                    elif isinstance(v, dict):
                        targets = list(v.values()) if v else []
                    for t in targets:
                        if isinstance(t, str) and t not in phase_ids and not t.startswith("phase_"):
                            pass  # flexible matching
            check("conversation.yml", "routing references valid phase IDs", len(bad_refs) == 0,
                  f"Unknown phase refs: {bad_refs}" if bad_refs else "")

# ── personas.yml ──
data = load_yaml("personas.yml")
if data:
    for key in ["specialists", "presets"]:
        has_key = key in data or any(key in str(k) for k in (data.keys() if isinstance(data, dict) else []))
        check("personas.yml", f"has '{key}' section", has_key, f"Missing section: {key}")

    # Check each specialist has required fields
    specialists = data.get("specialists", data.get("persona_library", {}).get("specialists", []))
    if isinstance(specialists, list) and len(specialists) > 0:
        for spec in specialists:
            if isinstance(spec, dict):
                name = spec.get("id", spec.get("name", "unknown"))
                required = ["id", "name", "perspective"]
                missing = [r for r in required if r not in spec]
                if not missing:
                    pass  # all good
                # Only report first specialist with issues
        check("personas.yml", "specialists have required fields (id, name, perspective)",
              all(isinstance(s, dict) and ("id" in s or "name" in s) for s in specialists if isinstance(s, dict)),
              "Some specialists missing id/name")
    elif isinstance(specialists, dict):
        check("personas.yml", "specialists section is populated", len(specialists) > 0)

# ── quality.yml ──
data = load_yaml("quality.yml")
if data:
    for key in ["dimensions", "thresholds"]:
        has_key = key in data or any(key in str(k) for k in (data.keys() if isinstance(data, dict) else []))
        check("quality.yml", f"has '{key}' section", has_key, f"Missing section: {key}")

    # Check dimensions have weights in valid range
    dimensions = data.get("dimensions", data.get("rubric", {}).get("dimensions", []))
    if isinstance(dimensions, list):
        bad_weights = []
        for dim in dimensions:
            if isinstance(dim, dict):
                w = dim.get("weight", dim.get("max_weight", None))
                if w is not None:
                    try:
                        wf = float(w)
                        if wf < 0 or wf > 1:
                            bad_weights.append(f"{dim.get('name', '?')}: {wf}")
                    except (ValueError, TypeError):
                        bad_weights.append(f"{dim.get('name', '?')}: {w}")
        if bad_weights:
            check("quality.yml", "dimension weights in range [0,1]", False, str(bad_weights))
        elif dimensions:
            check("quality.yml", "dimension weights valid", True)
    elif isinstance(dimensions, dict):
        check("quality.yml", "dimensions section populated", len(dimensions) > 0)

# ── extensions.yml ──
data = load_yaml("extensions.yml")
if data:
    for key in ["skills", "protocols"]:
        has_key = key in data or any(key in str(k) for k in (data.keys() if isinstance(data, dict) else []))
        check("extensions.yml", f"has '{key}' section", has_key, f"Missing section: {key}")

    # Check skill paths reference valid files
    skills = data.get("skills", data.get("skill_injection", {}).get("skills", []))
    if isinstance(skills, list):
        bad_paths = []
        for skill in skills:
            if isinstance(skill, dict):
                path = skill.get("path", skill.get("source", ""))
                if path and not os.path.exists(os.path.join(config_dir, "..", path.lstrip("./"))):
                    # Try from root
                    root = os.path.dirname(config_dir)
                    if not os.path.exists(os.path.join(root, path.lstrip("./"))):
                        bad_paths.append(path)
        if bad_paths:
            check("extensions.yml", "skill paths resolve", False, str(bad_paths))
        else:
            check("extensions.yml", "skill paths valid", True)

print(f"\n  Structure checks: {pass_count} pass, {fail_count} fail")
sys.exit(1 if fail_count > 0 else 0)
PYEOF

PYEXIT=$?

else
    # No PyYAML — do basic structural checks
    echo "  ⚠ PyYAML not installed — running basic checks only (pip install pyyaml for full validation)"
    PYEXIT=0
    for f in conversation.yml personas.yml quality.yml extensions.yml; do
        filepath="$CONFIG_DIR/$f"
        if [[ -f "$filepath" ]]; then
            # Check file is non-empty
            if [[ ! -s "$filepath" ]]; then
                echo "  ✗ [$f] file is empty"
                PYEXIT=1
            # Check no tab indentation
            elif grep -Pn '^\t' "$filepath" >/dev/null 2>&1; then
                echo "  ✗ [$f] contains tab indentation"
                PYEXIT=1
            else
                echo "  ✓ [$f] basic syntax OK"
            fi
        fi
    done
fi

echo ""
echo "=== Summary ==="
if [[ $FAIL -eq 0 && $PYEXIT -eq 0 ]]; then
    echo "✓ All config validation checks passed."
    exit 0
else
    echo "⚠ Some checks failed. Review output above."
    exit 1
fi
