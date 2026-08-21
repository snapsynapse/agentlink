#!/bin/sh
set -eu

targets_file="scripts/release-targets.txt"
test -s "$targets_file" || { echo "missing release targets" >&2; exit 1; }

count=0
seen=" "
while IFS= read -r target || test -n "$target"; do
  case "$target" in
    darwin/arm64|darwin/amd64|linux/amd64|linux/arm64) ;;
    *) echo "unsupported release target: $target" >&2; exit 1 ;;
  esac
  case "$seen" in
    *" $target "*) echo "duplicate release target: $target" >&2; exit 1 ;;
  esac
  seen="$seen$target "
  count=$((count + 1))
done < "$targets_file"

test "$count" -eq 4 || { echo "expected four release targets, found $count" >&2; exit 1; }

for script in scripts/prepare-release.sh scripts/release.sh scripts/verify-release.sh; do
  test -f "$script" || { echo "missing $script" >&2; exit 1; }
  grep -Fq "$targets_file" "$script" || { echo "$script does not consume $targets_file" >&2; exit 1; }
done

if grep -Eq 'git (commit|push origin HEAD:main)' scripts/release.sh; then
  echo "release.sh must not commit or push main" >&2
  exit 1
fi

echo "Release contract OK: $count targets, preparation separated from publication."
