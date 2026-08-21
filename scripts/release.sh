#!/bin/sh
# Publish a prepared Agentlink release and update snapsynapse/homebrew-tap.
set -eu

version="${1:?usage: scripts/release.sh <version, e.g. 1.2.3>}"
case "$version" in v*) version="${version#v}" ;; esac
tag="v$version"
repo="${REPO:-snapsynapse/agentlink}"
tap_repo="${TAP_REPO:-snapsynapse/homebrew-tap}"
targets_file="scripts/release-targets.txt"
notes_file="RELEASE_NOTES-$version.md"

test "$(git branch --show-current)" = "main" || { echo "release must run from main" >&2; exit 1; }
test -z "$(git status --porcelain)" || { echo "working tree not clean" >&2; exit 1; }
git fetch origin main --tags
test "$(git rev-parse HEAD)" = "$(git rev-parse origin/main)" || { echo "local main does not match origin/main" >&2; exit 1; }
if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
  echo "tag already exists: $tag" >&2
  exit 1
fi
if gh release view "$tag" --repo "$repo" >/dev/null 2>&1; then
  echo "GitHub Release already exists: $tag" >&2
  exit 1
fi

sh scripts/prepare-release.sh "$version"

set -- "dist/SHA256SUMS.txt" "$notes_file"
while IFS= read -r target || test -n "$target"; do
  set -- "$@" "dist/agentlink-${target%/*}-${target#*/}"
done < "$targets_file"

git tag -a "$tag" -m "Agentlink $tag"
git push origin "$tag"
gh release create "$tag" --repo "$repo" --latest \
  --target "$(git rev-parse HEAD)" \
  --title "Agentlink $tag" \
  --notes-file "$notes_file" \
  "$@"

sha() { awk -v f="agentlink-$1" '$2==f{print $1}' dist/SHA256SUMS.txt; }
tap_tmp="$(mktemp -d "${TMPDIR:-/tmp}/agentlink-tap.XXXXXX")"
trap 'rm -rf "$tap_tmp"' EXIT HUP INT TERM
gh repo clone "$tap_repo" "$tap_tmp/tap" -- --depth 1 -q
cat > "$tap_tmp/tap/Formula/agentlink.rb" <<EOF
class Agentlink < Formula
  desc "Sync one AGENTS.md to every AI coding tool - symlinks, no codegen"
  homepage "https://agentlink.run/"
  version "$version"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/$repo/releases/download/$tag/agentlink-darwin-arm64"
      sha256 "$(sha darwin-arm64)"
    else
      url "https://github.com/$repo/releases/download/$tag/agentlink-darwin-amd64"
      sha256 "$(sha darwin-amd64)"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/$repo/releases/download/$tag/agentlink-linux-arm64"
      sha256 "$(sha linux-arm64)"
    else
      url "https://github.com/$repo/releases/download/$tag/agentlink-linux-amd64"
      sha256 "$(sha linux-amd64)"
    end
  end

  def install
    bin.install Dir["agentlink-*"].first => "agentlink"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/agentlink --version")
  end
end
EOF
git -C "$tap_tmp/tap" add Formula/agentlink.rb
git -C "$tap_tmp/tap" commit -qm "agentlink $version"
git -C "$tap_tmp/tap" push -q

sh scripts/verify-release.sh "$tag"
echo "Published and verified Agentlink $tag."
