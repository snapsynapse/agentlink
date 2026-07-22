#!/bin/sh
# Release agentlink: build binaries, tag, publish the GitHub release, and
# update the Homebrew tap formula in snapsynapse/homebrew-tap.
#
# Usage: scripts/release.sh 0.4.0
#
# Expects: clean working tree, CHANGELOG.md already has the version section,
# gh authenticated, and push access to snapsynapse/homebrew-tap.
set -eu

version="${1:?usage: scripts/release.sh <version, e.g. 0.4.0>}"
case "$version" in v*) version="${version#v}" ;; esac
tag="v$version"
repo="${REPO:-snapsynapse/agentlink}"
tap_repo="${TAP_REPO:-snapsynapse/homebrew-tap}"
module="github.com/snapsynapse/agentlink"
notes_file="RELEASE_NOTES-$version.md"

test -z "$(git status --porcelain)" || { echo "working tree not clean" >&2; exit 1; }
grep -Fq "[$version]" CHANGELOG.md || { echo "CHANGELOG.md has no [$version] section" >&2; exit 1; }
test -f "$notes_file" || { echo "missing $notes_file" >&2; exit 1; }

echo "==> Tests"
go mod tidy -diff
go vet ./...
go test ./...
go test -race ./...
go test -tags=integration .

echo "==> Build $tag"
rm -rf dist && mkdir -p dist
for target in darwin/arm64 darwin/amd64 linux/amd64 linux/arm64; do
  goos="${target%/*}" goarch="${target#*/}"
  GOOS="$goos" GOARCH="$goarch" go build -trimpath \
    -ldflags "-s -w -X $module/internal/cli.version=$version" \
    -o "dist/agentlink-$goos-$goarch" ./cmd/agentlink
done
(cd dist && shasum -a 256 agentlink-* > SHA256SUMS.txt)

echo "==> Bump landing page"
prev_tag="$(git describe --tags --abbrev=0)"
if [ "$prev_tag" != "$tag" ] && grep -Fq "$prev_tag" docs/index.html; then
  sed -i '' "s|$prev_tag|$tag|g" docs/index.html
  git add docs/index.html
  git commit -qm "Bump landing page to $tag"
fi

echo "==> Tag and release"
git tag -a "$tag" -m "Agentlink $tag"
git push origin HEAD:main "$tag"
gh release create "$tag" --repo "$repo" --latest \
  --title "Agentlink $tag" \
  --notes-file "$notes_file" \
  dist/agentlink-darwin-arm64 dist/agentlink-darwin-amd64 \
  dist/agentlink-linux-amd64 dist/agentlink-linux-arm64 dist/SHA256SUMS.txt \
  "$notes_file"

echo "==> Update Homebrew tap"
sha() { awk -v f="agentlink-$1" '$2==f{print $1}' dist/SHA256SUMS.txt; }
tmp="$(mktemp -d)"
gh repo clone "$tap_repo" "$tmp/tap" -- --depth 1 -q
cat > "$tmp/tap/Formula/agentlink.rb" <<EOF
class Agentlink < Formula
  desc "Sync one AGENTS.md to every AI coding tool — symlinks, no codegen"
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
git -C "$tmp/tap" commit -aqm "agentlink $version"
git -C "$tmp/tap" push -q
rm -rf "$tmp"

echo "==> Verify"
sh scripts/verify-release.sh "$tag"
echo "Done."
