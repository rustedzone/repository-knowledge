#!/bin/sh

set -eu

fail() {
  printf 'error: %s\n' "$*" >&2
  exit 2
}

binary=${REPO_KNOWLEDGE_BINARY:-}
target=${REPO_KNOWLEDGE_TARGET:-${GITHUB_WORKSPACE:-}}
output=${REPO_KNOWLEDGE_OUTPUT:-${RUNNER_TEMP:-/tmp}/repository-knowledge-impact.json}
mode=${REPO_KNOWLEDGE_ENFORCEMENT:-advisory}
base=${REPO_KNOWLEDGE_BASE_SHA:-}
head=${REPO_KNOWLEDGE_HEAD_SHA:-${GITHUB_SHA:-}}

[ -n "$binary" ] || fail "REPO_KNOWLEDGE_BINARY is required"
[ -x "$binary" ] || fail "repo-knowledge binary is not executable: $binary"
[ -n "$target" ] || fail "REPO_KNOWLEDGE_TARGET or GITHUB_WORKSPACE is required"
[ -d "$target" ] || fail "target repository does not exist: $target"

case "$mode" in
  advisory|acknowledgment|enforced) ;;
  *) fail "unsupported enforcement mode: $mode" ;;
esac

is_object_id() {
  printf '%s\n' "$1" | grep -Eq '^([0-9a-fA-F]{40}|[0-9a-fA-F]{64})$'
}

is_zero_object_id() {
  printf '%s\n' "$1" | grep -Eq '^(0{40}|0{64})$'
}

if [ -z "$base" ] || is_zero_object_id "$base"; then
  base=$(git -C "$target" hash-object -t tree /dev/null)
elif ! is_object_id "$base"; then
  fail "base must be a Git object ID: $base"
fi

[ -n "$head" ] || fail "REPO_KNOWLEDGE_HEAD_SHA or GITHUB_SHA is required"
is_object_id "$head" || fail "head must be a Git object ID: $head"

ensure_tree() {
  object=$1
  if git -C "$target" cat-file -e "${object}^{tree}" 2>/dev/null; then
    return 0
  fi
  fail "Git object is unavailable in the full checkout: $object"
}

ensure_tree "$base"
ensure_tree "$head"
mkdir -p "$(dirname "$output")"

set +e
"$binary" validate-doc-impact \
  --target "$target" \
  --base "$base" \
  --head "$head" \
  --mode "$mode" \
  --json > "$output"
validation_status=$?
set -e

[ -s "$output" ] || fail "validator produced no JSON report"
cat "$output"
exit "$validation_status"
