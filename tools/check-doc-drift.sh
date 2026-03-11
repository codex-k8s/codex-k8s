#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  tools/check-doc-drift.sh [--issues <issue-number>...]

Checks:
  - stale repo-local references such as docs/README.md or docs/03_engineering/*
  - broken repo-relative docs/services.yaml path references in tracked markdown/yaml files
  - stale same-repo blob links in tracked markdown/yaml files
  - optional GitHub issue bodies passed via --issues
EOF
}

issue_numbers=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --issues)
      shift
      while [[ $# -gt 0 && "$1" != --* ]]; do
        issue_numbers+=("$1")
        shift
      done
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

failures=0
migration_map_path="docs/delivery/documentation_ia_migration_map.md"

tracked_sources="$tmp_dir/tracked_sources.txt"
git ls-files '*.md' '*.yaml' '*.yml' > "$tracked_sources"
printf '%s\n' services.yaml >> "$tracked_sources"
sort -u "$tracked_sources" -o "$tracked_sources"

checked_sources="$tmp_dir/checked_sources.txt"
grep -Ev '^docs/delivery/(epics|sprints)/|^docs/architecture/(adr|alternatives)/' "$tracked_sources" > "$checked_sources"

extract_repo_paths() {
  local input_file="$1"
  grep -oE '`(docs/[A-Za-z0-9_./-]+\.md|services\.yaml)`' "$input_file" \
    | tr -d '`' \
    | sort -u
}

check_missing_paths() {
  local source_name="$1"
  local input_file="$2"
  local missing=0
  while IFS= read -r ref; do
    [[ -z "$ref" ]] && continue
    if [[ ! -e "$ref" ]]; then
      echo "missing path in ${source_name}: ${ref}" >&2
      missing=1
    fi
  done < <(extract_repo_paths "$input_file")
  if [[ $missing -ne 0 ]]; then
    failures=1
  fi
}

check_forbidden_patterns() {
  local source_name="$1"
  local input_file="$2"
  if grep -nE 'docs/README\.md|docs/03_engineering/' "$input_file" >/dev/null; then
    echo "stale legacy docs path in ${source_name}" >&2
    grep -nE 'docs/README\.md|docs/03_engineering/' "$input_file" >&2
    failures=1
  fi
  if grep -nE 'https://github\.com/codex-k8s/codex-k8s/blob/' "$input_file" >/dev/null; then
    echo "stale same-repo blob link in ${source_name}" >&2
    grep -nE 'https://github\.com/codex-k8s/codex-k8s/blob/' "$input_file" >&2
    failures=1
  fi
}

while IFS= read -r file; do
  [[ -z "$file" ]] && continue
  [[ -f "$file" ]] || continue
  if [[ "$file" == "$migration_map_path" ]]; then
    continue
  fi
  check_forbidden_patterns "$file" "$file"
  check_missing_paths "$file" "$file"
done < "$checked_sources"

services_paths="$tmp_dir/services_paths.txt"
rg -No '^\s*-?\s*path:\s+\S+$' services.yaml \
  | sed -E 's/^\s*-?\s*path:\s+//' \
  | sort -u > "$services_paths"

while IFS= read -r path_ref; do
  [[ -z "$path_ref" ]] && continue
  if [[ ! -e "$path_ref" ]]; then
    echo "services.yaml references missing path: ${path_ref}" >&2
    failures=1
  fi
done < "$services_paths"

if [[ ${#issue_numbers[@]} -gt 0 ]]; then
  if ! command -v gh >/dev/null 2>&1; then
    echo "gh is required when --issues is used" >&2
    exit 1
  fi

  for issue_number in "${issue_numbers[@]}"; do
    issue_body_file="$tmp_dir/issue-${issue_number}.md"
    gh issue view "$issue_number" --json body --jq .body > "$issue_body_file"
    check_forbidden_patterns "issue #${issue_number}" "$issue_body_file"
    check_missing_paths "issue #${issue_number}" "$issue_body_file"
  done
fi

if [[ $failures -ne 0 ]]; then
  exit 1
fi

echo "doc-drift check passed"
