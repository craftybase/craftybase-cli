#!/bin/sh
# Verifies the installer resolves GoReleaser asset names without network I/O.
set -eu
out=$(STOCKSMITH_DRY_RUN=1 STOCKSMITH_OS=darwin STOCKSMITH_ARCH=arm64 STOCKSMITH_VERSION=v1.2.3 \
  sh website/public/install)
archive=$(printf '%s\n' "$out" | sed -n 1p)
url=$(printf '%s\n' "$out" | sed -n 2p)
[ "$archive" = "stocksmith_1.2.3_darwin_arm64.tar.gz" ] || { echo "bad archive: $archive" >&2; exit 1; }
case "$url" in
  https://github.com/craftybase/stocksmith-cli/releases/download/v1.2.3/stocksmith_1.2.3_darwin_arm64.tar.gz) ;;
  *) echo "bad url: $url" >&2; exit 1 ;;
esac
echo OK
