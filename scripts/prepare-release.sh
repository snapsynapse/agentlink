#!/bin/sh
# Build and verify a release candidate without publishing it.
set -eu

version="${1:?usage: scripts/prepare-release.sh <version, e.g. 1.2.3>}"
case "$version" in v*) version="${version#v}" ;; esac
case "$version" in
  *[!0-9.]*|*.*.*.*|.*|*.) echo "version must be X.Y.Z" >&2; exit 1 ;;
esac
test "$(printf '%s' "$version" | awk -F. '{print NF}')" -eq 3 || { echo "version must be X.Y.Z" >&2; exit 1; }

module="github.com/snapsynapse/agentlink"
targets_file="scripts/release-targets.txt"
notes_file="RELEASE_NOTES-$version.md"

test -z "$(git status --porcelain)" || { echo "working tree not clean" >&2; exit 1; }
grep -Fq "[$version]" CHANGELOG.md || { echo "CHANGELOG.md has no [$version] section" >&2; exit 1; }
test -f "$notes_file" || { echo "missing $notes_file" >&2; exit 1; }
sh scripts/check-release-contract.sh

go mod tidy -diff
go vet ./...
go test ./...
go test -race ./...
go test -tags=integration ./...

rm -rf dist
mkdir -p dist
while IFS= read -r target || test -n "$target"; do
  goos="${target%/*}"
  goarch="${target#*/}"
  GOOS="$goos" GOARCH="$goarch" go build -trimpath \
    -ldflags "-s -w -X $module/internal/cli.version=$version" \
    -o "dist/agentlink-$goos-$goarch" ./cmd/agentlink
done < "$targets_file"
(cd dist && shasum -a 256 agentlink-* > SHA256SUMS.txt)

rebuild_dir="$(mktemp -d "${TMPDIR:-/tmp}/agentlink-rebuild.XXXXXX")"
trap 'rm -rf "$rebuild_dir"' EXIT HUP INT TERM
while IFS= read -r target || test -n "$target"; do
  goos="${target%/*}"
  goarch="${target#*/}"
  output="agentlink-$goos-$goarch"
  GOOS="$goos" GOARCH="$goarch" go build -trimpath \
    -ldflags "-s -w -X $module/internal/cli.version=$version" \
    -o "$rebuild_dir/$output" ./cmd/agentlink
  cmp "dist/$output" "$rebuild_dir/$output" || { echo "non-deterministic rebuild: $output" >&2; exit 1; }
done < "$targets_file"

case "$(uname -s)/$(uname -m)" in
  Darwin/arm64) host_asset="dist/agentlink-darwin-arm64" ;;
  Darwin/x86_64) host_asset="dist/agentlink-darwin-amd64" ;;
  Linux/aarch64|Linux/arm64) host_asset="dist/agentlink-linux-arm64" ;;
  Linux/x86_64) host_asset="dist/agentlink-linux-amd64" ;;
  *) host_asset="" ;;
esac
if test -n "$host_asset"; then
  test "$("$host_asset" --version)" = "agentlink version $version" || { echo "host binary version mismatch" >&2; exit 1; }
fi

echo "Prepared deterministic Agentlink v$version assets in dist/."
