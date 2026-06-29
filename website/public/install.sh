#!/bin/sh
# stocksmith installer — https://cli.stocksmith.dev/install
# Downloads the matching GoReleaser release asset, verifies its SHA-256, and
# installs the `stocksmith` binary.
#
# Env overrides:
#   STOCKSMITH_VERSION       specific version, e.g. v1.2.3 (default: latest release)
#   STOCKSMITH_INSTALL_DIR   install dir (default: /usr/local/bin, else ~/.local/bin)
#   STOCKSMITH_OS / _ARCH    force platform (mainly for testing)
#   STOCKSMITH_DRY_RUN=1     print "<archive>\n<asset-url>" and exit (no download)
#
# Asset names MUST match .goreleaser.yml name_template:
#   stocksmith_<version>_<os>_<arch>.tar.gz  and  stocksmith_<version>_checksums.txt
set -eu

REPO="craftybase/stocksmith-cli"
BINARY="stocksmith"

err()  { printf 'error: %s\n' "$1" >&2; exit 1; }
info() { printf '%s\n' "$1" >&2; }
have() { command -v "$1" >/dev/null 2>&1; }

download_stdout() { if have curl; then curl -fsSL "$1"; elif have wget; then wget -qO- "$1"; else err "need curl or wget"; fi; }
download_file()   { if have curl; then curl -fsSL "$1" -o "$2"; elif have wget; then wget -qO "$2" "$1"; else err "need curl or wget"; fi; }
sha256_of()       { if have sha256sum; then sha256sum "$1" | awk '{print $1}'; elif have shasum; then shasum -a 256 "$1" | awk '{print $1}'; else err "need sha256sum or shasum"; fi; }

detect_os() {
  [ -n "${STOCKSMITH_OS:-}" ] && { printf '%s' "$STOCKSMITH_OS"; return; }
  case "$(uname -s)" in
    Linux)  printf 'linux' ;;
    Darwin) printf 'darwin' ;;
    MINGW*|MSYS*|CYGWIN*) err "Windows isn't supported by this installer. Use WSL, or download a release zip from https://github.com/$REPO/releases. See https://cli.stocksmith.dev/getting-started/." ;;
    *) err "unsupported OS '$(uname -s)'. Download a release from https://github.com/$REPO/releases or install via Homebrew. See https://cli.stocksmith.dev/getting-started/." ;;
  esac
}
detect_arch() {
  [ -n "${STOCKSMITH_ARCH:-}" ] && { printf '%s' "$STOCKSMITH_ARCH"; return; }
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
VERSION=${STOCKSMITH_VERSION:-$(latest_version)}
VNUM=${VERSION#v}
ARCHIVE="${BINARY}_${VNUM}_${OS}_${ARCH}.tar.gz"
CHECKSUMS="${BINARY}_${VNUM}_checksums.txt"
BASE="https://github.com/$REPO/releases/download/$VERSION"
ASSET_URL="$BASE/$ARCHIVE"

if [ "${STOCKSMITH_DRY_RUN:-}" = "1" ]; then
  printf '%s\n%s\n' "$ARCHIVE" "$ASSET_URL"
  exit 0
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
info "Downloading $ARCHIVE ($VERSION, $OS/$ARCH)…"
download_file "$ASSET_URL" "$tmp/$ARCHIVE"
download_file "$BASE/$CHECKSUMS" "$tmp/$CHECKSUMS"

info "Verifying checksum…"
want=$(grep -F " $ARCHIVE" "$tmp/$CHECKSUMS" | awk '{print $1}')
[ -n "$want" ] || err "no checksum for $ARCHIVE"
got=$(sha256_of "$tmp/$ARCHIVE")
[ "$want" = "$got" ] || err "checksum mismatch (want $want, got $got)"

tar -xzf "$tmp/$ARCHIVE" -C "$tmp"
[ -f "$tmp/$BINARY" ] || err "archive did not contain $BINARY"
chmod +x "$tmp/$BINARY"

dir=${STOCKSMITH_INSTALL_DIR:-}
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
