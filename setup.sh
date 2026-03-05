#!/usr/bin/env bash

set -euo pipefail

OLD_MODULE="github.com/ulbwa/go-backend-template"

usage() {
  cat <<EOF
Usage: $(basename "$0") [NEW_MODULE]

Replaces module path references:
  $OLD_MODULE -> NEW_MODULE

Targets:
  - go.mod
  - go.sum
  - all .go files (excluding ./vendor)

Behavior:
  1) Dry-run preview (no changes)
  2) Confirmation prompt
  3) Apply changes only after confirmation
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -gt 1 ]]; then
  usage
  exit 1
fi

NEW_MODULE="${1:-}"
if [[ -z "$NEW_MODULE" ]]; then
  read -r -p "Enter new module path: " NEW_MODULE
fi

if [[ -z "$NEW_MODULE" ]]; then
  echo "Error: new module path is required." >&2
  exit 1
fi

if [[ "$NEW_MODULE" == "$OLD_MODULE" ]]; then
  echo "Error: new module path is the same as current one." >&2
  exit 1
fi

if [[ ! -f "go.mod" ]]; then
  echo "Error: go.mod not found. Run this script from project root." >&2
  exit 1
fi

declare -a files=()
files+=("go.mod")
[[ -f "go.sum" ]] && files+=("go.sum")

while IFS= read -r -d '' file; do
  files+=("$file")
done < <(find . -type f -name '*.go' ! -path './vendor/*' -print0 | sort -z)

declare -a target_files=()
declare -A counts=()
total_matches=0

for file in "${files[@]}"; do
  [[ -f "$file" ]] || continue
  match_count=$(grep -F -c "$OLD_MODULE" "$file" || true)
  if (( match_count > 0 )); then
    target_files+=("$file")
    counts["$file"]=$match_count
    (( total_matches += match_count ))
  fi
done

if (( total_matches == 0 )); then
  echo "No occurrences found for: $OLD_MODULE"
  echo "Nothing to change."
  exit 0
fi

echo "Dry-run preview (no files changed):"
echo "  from: $OLD_MODULE"
echo "  to:   $NEW_MODULE"
echo
for file in "${target_files[@]}"; do
  echo "  - $file (${counts[$file]} matches)"
done
echo
echo "Total replacements: $total_matches"

read -r -p "Apply these changes? [y/N]: " confirm
case "${confirm,,}" in
  y|yes)
    ;;
  *)
    echo "Aborted. No files were changed."
    exit 0
    ;;
esac

escaped_old=$(printf '%s' "$OLD_MODULE" | sed -e 's/[\/&|]/\\&/g')
escaped_new=$(printf '%s' "$NEW_MODULE" | sed -e 's/[\/&|]/\\&/g')

for file in "${target_files[@]}"; do
  tmp_file=$(mktemp)
  sed "s|$escaped_old|$escaped_new|g" "$file" > "$tmp_file"
  chmod --reference="$file" "$tmp_file" || true
  mv "$tmp_file" "$file"
  echo "Updated: $file"
done

echo
echo "Done. Updated ${#target_files[@]} file(s)."

script_path="${BASH_SOURCE[0]}"
if rm -- "$script_path"; then
  echo "Removed script: $script_path"
else
  echo "Warning: failed to remove script: $script_path" >&2
fi
