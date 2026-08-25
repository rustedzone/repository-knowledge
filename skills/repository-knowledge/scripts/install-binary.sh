#!/bin/sh

set -eu

repository="rustedzone/repository-knowledge"
version=""
bin_dir=""
assume_yes=false
dry_run=false

usage() {
  cat <<'EOF'
Install a pinned repo-knowledge release into a user-local directory on PATH.

Usage:
  install-binary.sh --version vX.Y.Z [--repository OWNER/REPO]
                    [--bin-dir /absolute/path] [--yes] [--dry-run]

The script never selects "latest", uses sudo, or edits shell startup files.
Without --yes it displays the exact operation and asks for confirmation.
EOF
}

fail() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

path_contains() {
  case ":${PATH:-}:" in
    *":$1:"*) return 0 ;;
    *) return 1 ;;
  esac
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      [ "$#" -ge 2 ] || fail "--version requires a value"
      version=$2
      shift 2
      ;;
    --repository)
      [ "$#" -ge 2 ] || fail "--repository requires a value"
      repository=$2
      shift 2
      ;;
    --bin-dir)
      [ "$#" -ge 2 ] || fail "--bin-dir requires a value"
      bin_dir=$2
      shift 2
      ;;
    --yes)
      assume_yes=true
      shift
      ;;
    --dry-run)
      dry_run=true
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *) fail "unknown argument: $1" ;;
  esac
done

if ! printf '%s\n' "$version" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$'; then
  fail "--version must be a pinned semantic release tag such as v0.8.0"
fi
case "$repository" in
  *[!A-Za-z0-9_.\/-]*|/*|*/|*//*|*/*/*|"") fail "--repository must be OWNER/REPO" ;;
esac

os=$(uname -s)
case "$os" in
  Darwin) platform=darwin ;;
  Linux) platform=linux ;;
  *) fail "unsupported operating system: $os" ;;
esac

machine=$(uname -m)
case "$machine" in
  x86_64|amd64) architecture=amd64 ;;
  arm64|aarch64) architecture=arm64 ;;
  *) fail "unsupported architecture: $machine" ;;
esac

artifact="repo-knowledge-${platform}-${architecture}"
release_url="https://github.com/${repository}/releases/download/${version}"

if [ -z "$bin_dir" ]; then
  if path_contains "$HOME/.local/bin"; then
    bin_dir="$HOME/.local/bin"
  elif path_contains "$HOME/bin"; then
    bin_dir="$HOME/bin"
  else
    fail "neither $HOME/.local/bin nor $HOME/bin is on PATH; add one to PATH or pass --bin-dir with a user-owned directory already on PATH"
  fi
fi

case "$bin_dir" in
  /*) ;;
  *) fail "--bin-dir must be an absolute path" ;;
esac
path_contains "$bin_dir" || fail "destination is not on PATH: $bin_dir"

destination="$bin_dir/repo-knowledge"
printf '%s\n' \
  "Pinned release: $version" \
  "Source:         https://github.com/$repository" \
  "Artifact:       $artifact" \
  "Destination:    $destination" \
  "Integrity:      SHA256SUMS; GitHub attestation too when a compatible gh command is available"

if [ "$dry_run" = true ]; then
  exit 0
fi

if [ "$assume_yes" != true ]; then
  [ -t 0 ] || fail "confirmation required; review the plan and rerun with --yes"
  printf 'Install or replace this executable? [y/N] '
  read -r answer
  case "$answer" in
    y|Y|yes|YES) ;;
    *) fail "installation cancelled" ;;
  esac
fi

command -v curl >/dev/null 2>&1 || fail "curl is required"
if command -v shasum >/dev/null 2>&1; then
  checksum_command=shasum
elif command -v sha256sum >/dev/null 2>&1; then
  checksum_command=sha256sum
else
  fail "shasum or sha256sum is required"
fi

temporary_dir=$(mktemp -d "${TMPDIR:-/tmp}/repo-knowledge-install.XXXXXX")
cleanup() {
  rm -rf "$temporary_dir"
}
trap cleanup EXIT HUP INT TERM

curl --proto '=https' --tlsv1.2 --fail --location --silent --show-error \
  --output "$temporary_dir/$artifact" "$release_url/$artifact"
curl --proto '=https' --tlsv1.2 --fail --location --silent --show-error \
  --output "$temporary_dir/SHA256SUMS" "$release_url/SHA256SUMS"

expected=$(awk -v artifact="$artifact" '$2 == artifact { print $1 }' "$temporary_dir/SHA256SUMS")
[ -n "$expected" ] || fail "SHA256SUMS has no entry for $artifact"
[ "${#expected}" -eq 64 ] || fail "invalid SHA-256 entry for $artifact"
case "$expected" in
  *[!0-9a-fA-F]*) fail "invalid SHA-256 entry for $artifact" ;;
esac

if [ "$checksum_command" = shasum ]; then
  actual=$(shasum -a 256 "$temporary_dir/$artifact" | awk '{ print $1 }')
else
  actual=$(sha256sum "$temporary_dir/$artifact" | awk '{ print $1 }')
fi
[ "$actual" = "$expected" ] || fail "checksum mismatch for $artifact"

chmod 0755 "$temporary_dir/$artifact"
reported_version=$("$temporary_dir/$artifact" --version) || fail "downloaded artifact could not run"
[ "$reported_version" = "repo-knowledge ${version#v}" ] || fail "downloaded artifact reports unexpected version: $reported_version"

if command -v gh >/dev/null 2>&1 && gh attestation verify --help >/dev/null 2>&1; then
  gh attestation verify "$temporary_dir/$artifact" --repo "$repository" >/dev/null
  printf 'Verified GitHub artifact attestation.\n'
else
  printf 'GitHub attestation verification skipped because a compatible gh command is unavailable.\n'
fi

mkdir -p "$bin_dir"
install -m 0755 "$temporary_dir/$artifact" "$destination"
printf 'Installed repo-knowledge at %s (available as repo-knowledge on the current PATH).\n' "$destination"
