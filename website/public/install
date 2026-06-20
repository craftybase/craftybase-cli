#!/bin/sh
# craftybase installer — https://cli.craftybase.dev/install
# Downloads the matching GoReleaser release asset, verifies its SHA-256, and
# installs the `craftybase` binary.
#
# Env overrides:
#   CRAFTYBASE_VERSION       specific version, e.g. v1.2.3 (default: latest release)
#   CRAFTYBASE_INSTALL_DIR   install dir (default: /usr/local/bin, else ~/.local/bin)
#   CRAFTYBASE_OS / _ARCH    force platform (mainly for testing)
#   CRAFTYBASE_DRY_RUN=1     print "<archive>\n<asset-url>" and exit (no download)
#
# Asset names MUST match .goreleaser.yml name_template:
#   craftybase_<version>_<os>_<arch>.tar.gz  and  craftybase_<version>_checksums.txt
set -eu

REPO="craftybase/craftybase-cli"
BINARY="craftybase"

err()  { printf 'error: %s\n' "$1" >&2; exit 1; }
info() { printf '%s\n' "$1" >&2; }
have() { command -v "$1" >/dev/null 2>&1; }

download_stdout() { if have curl; then curl -fsSL "$1"; elif have wget; then wget -qO- "$1"; else err "need curl or wget"; fi; }
download_file()   { if have curl; then curl -fsSL "$1" -o "$2"; elif have wget; then wget -qO "$2" "$1"; else err "need curl or wget"; fi; }
sha256_of()       { if have sha256sum; then sha256sum "$1" | awk '{print $1}'; elif have shasum; then shasum -a 256 "$1" | awk '{print $1}'; else err "need sha256sum or shasum"; fi; }

detect_os() {
  [ -n "${CRAFTYBASE_OS:-}" ] && { printf '%s' "$CRAFTYBASE_OS"; return; }
  case "$(uname -s)" in
    Linux)  printf 'linux' ;;
    Darwin) printf 'darwin' ;;
    *) err "unsupported OS '$(uname -s)'. On Windows, download a release from https://github.com/$REPO/releases or use WSL. See https://cli.craftybase.dev/getting-started/." ;;
  esac
}
detect_arch() {
  [ -n "${CRAFTYBASE_ARCH:-}" ] && { printf '%s' "$CRAFTYBASE_ARCH"; return; }
  case "$(uname -m)" in
    x86_64|amd64)  printf 'amd64' ;;
    arm64|aarch64) printf 'arm64' ;;
    *) err "unsupported architecture '$(uname -m)'." ;;
  esac
}
latest_version() {
  tag=$(download_stdout "https://api.github.com/repos/$REPO/releases/latest" \
        | grep '"tag_name"' | head -n1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
  [ -n "$tag" ] || err "could not determine latest version"
  printf '%s' "$tag"
}

OS=$(detect_os)
ARCH=$(detect_arch)
VERSION=${CRAFTYBASE_VERSION:-$(latest_version)}
VNUM=${VERSION#v}
ARCHIVE="${BINARY}_${VNUM}_${OS}_${ARCH}.tar.gz"
CHECKSUMS="${BINARY}_${VNUM}_checksums.txt"
BASE="https://github.com/$REPO/releases/download/$VERSION"
ASSET_URL="$BASE/$ARCHIVE"

if [ "${CRAFTYBASE_DRY_RUN:-}" = "1" ]; then
  printf '%s\n%s\n' "$ARCHIVE" "$ASSET_URL"
  exit 0
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
info "Downloading $ARCHIVE ($VERSION, $OS/$ARCH)…"
download_file "$ASSET_URL" "$tmp/$ARCHIVE"
download_file "$BASE/$CHECKSUMS" "$tmp/$CHECKSUMS"

info "Verifying checksum…"
want=$(grep " $ARCHIVE\$" "$tmp/$CHECKSUMS" | awk '{print $1}')
[ -n "$want" ] || err "no checksum for $ARCHIVE"
got=$(sha256_of "$tmp/$ARCHIVE")
[ "$want" = "$got" ] || err "checksum mismatch (want $want, got $got)"

tar -xzf "$tmp/$ARCHIVE" -C "$tmp"
[ -f "$tmp/$BINARY" ] || err "archive did not contain $BINARY"
chmod +x "$tmp/$BINARY"

dir=${CRAFTYBASE_INSTALL_DIR:-}
if [ -z "$dir" ]; then
  if [ -w /usr/local/bin ]; then dir=/usr/local/bin; else dir="$HOME/.local/bin"; fi
fi
mkdir -p "$dir"
mv "$tmp/$BINARY" "$dir/$BINARY"
info "Installed $BINARY to $dir/$BINARY"
case ":$PATH:" in
  *":$dir:"*) ;;
  *) info "note: $dir is not on PATH — add: export PATH=\"$dir:\$PATH\"" ;;
esac
info "Next: run '$BINARY auth login'"
