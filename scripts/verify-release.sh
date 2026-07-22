#!/bin/sh
set -eu

version="${1:?usage: scripts/verify-release.sh <version, e.g. v1.2.3>}"
repo="${REPO:-snapsynapse/agentlink}"

case "$version" in
  v*) ;;
  *) version="v$version" ;;
esac

tmp="$(mktemp -d "${TMPDIR:-/tmp}/agentlink-release-${version}.XXXXXX")"

gh release view "$version" --repo "$repo" --json tagName,url,assets >"$tmp/release-view.json"
gh release download "$version" --repo "$repo" --dir "$tmp" --clobber

for asset in agentlink-darwin-arm64 agentlink-darwin-amd64 agentlink-linux-amd64 agentlink-linux-arm64 SHA256SUMS.txt "RELEASE_NOTES-${version#v}.md"; do
  test -f "$tmp/$asset"
done

(cd "$tmp" && shasum -a 256 -c SHA256SUMS.txt)

release_url="https://github.com/$repo/releases/tag/$version"
download_url="https://github.com/$repo/releases/download/$version/agentlink-darwin-arm64"

grep -F "$release_url" docs/index.html >/dev/null
grep -F "$download_url" docs/index.html >/dev/null

echo "Verified $release_url"
