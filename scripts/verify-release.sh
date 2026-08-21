#!/bin/sh
set -eu

version="${1:?usage: scripts/verify-release.sh <version, e.g. v1.2.3>}"
repo="${REPO:-snapsynapse/agentlink}"
targets_file="scripts/release-targets.txt"

case "$version" in v*) ;; *) version="v$version" ;; esac

tmp="$(mktemp -d "${TMPDIR:-/tmp}/agentlink-release-${version}.XXXXXX")"
trap 'rm -rf "$tmp"' EXIT HUP INT TERM

gh release view "$version" --repo "$repo" --json tagName,url,assets > "$tmp/release-view.json"
gh release download "$version" --repo "$repo" --dir "$tmp" --clobber

while IFS= read -r target || test -n "$target"; do
  test -f "$tmp/agentlink-${target%/*}-${target#*/}"
done < "$targets_file"
test -f "$tmp/SHA256SUMS.txt"
test -f "$tmp/RELEASE_NOTES-${version#v}.md"

(cd "$tmp" && shasum -a 256 -c SHA256SUMS.txt)

case "$(uname -s)/$(uname -m)" in
  Darwin/arm64) host_asset="$tmp/agentlink-darwin-arm64" ;;
  Darwin/x86_64) host_asset="$tmp/agentlink-darwin-amd64" ;;
  Linux/aarch64|Linux/arm64) host_asset="$tmp/agentlink-linux-arm64" ;;
  Linux/x86_64) host_asset="$tmp/agentlink-linux-amd64" ;;
  *) host_asset="" ;;
esac
if test -n "$host_asset"; then
  chmod +x "$host_asset"
  test "$("$host_asset" --version)" = "agentlink version ${version#v}" || { echo "published binary version mismatch" >&2; exit 1; }
fi

release_url="https://github.com/$repo/releases/tag/$version"
download_url="https://github.com/$repo/releases/download/$version/agentlink-darwin-arm64"
grep -F "$release_url" docs/index.html >/dev/null
grep -F "$download_url" docs/index.html >/dev/null

echo "Verified $release_url"
