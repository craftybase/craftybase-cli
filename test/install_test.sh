#!/bin/sh
# Verifies the installer resolves GoReleaser asset names without network I/O.
set -eu
out=$(CRAFTYBASE_DRY_RUN=1 CRAFTYBASE_OS=darwin CRAFTYBASE_ARCH=arm64 CRAFTYBASE_VERSION=v1.2.3 \
  sh website/public/install)
archive=$(printf '%s\n' "$out" | sed -n 1p)
url=$(printf '%s\n' "$out" | sed -n 2p)
[ "$archive" = "craftybase_1.2.3_darwin_arm64.tar.gz" ] || { echo "bad archive: $archive" >&2; exit 1; }
case "$url" in
  https://github.com/craftybase/craftybase-cli/releases/download/v1.2.3/craftybase_1.2.3_darwin_arm64.tar.gz) ;;
  *) echo "bad url: $url" >&2; exit 1 ;;
esac
echo OK
