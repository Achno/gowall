#!/usr/bin/env bash
set -euo pipefail

# Build a Debian (.deb) package for gowall from a compiled binary.
# Usage: ./scripts/build-deb.sh <binary-path> <version> <arch> <output-dir>

BINARY_PATH="${1:-gowall}"
VERSION="${2:-0.2.4}"
ARCH="${3:-amd64}"
OUTPUT_DIR="${4:-release}"

# Strip leading 'v' from version if present
VERSION="${VERSION#v}"

if [ ! -f "${BINARY_PATH}" ]; then
  echo "Error: binary '${BINARY_PATH}' not found" >&2
  exit 1
fi

case "${ARCH}" in
  amd64|x86_64)
    DEB_ARCH="amd64"
    ;;
  arm64|aarch64)
    DEB_ARCH="arm64"
    ;;
  *)
    DEB_ARCH="${ARCH}"
    ;;
esac

PKG_NAME="gowall"
PKG_DIR="$(mktemp -d)/${PKG_NAME}_${VERSION}_${DEB_ARCH}"

mkdir -p "${PKG_DIR}/DEBIAN"
mkdir -p "${PKG_DIR}/usr/bin"
mkdir -p "${OUTPUT_DIR}"

cp "${BINARY_PATH}" "${PKG_DIR}/usr/bin/gowall"
chmod 755 "${PKG_DIR}/usr/bin/gowall"

INSTALLED_SIZE=$(du -sk "${PKG_DIR}/usr" | cut -f1)

cat << EOF > "${PKG_DIR}/DEBIAN/control"
Package: ${PKG_NAME}
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${DEB_ARCH}
Installed-Size: ${INSTALLED_SIZE}
Maintainer: Adam Magued <adamismailmageud@gmail.com>
Homepage: https://github.com/Achno/gowall
Description: Wallpaper color scheme and palette converter tool
 Gowall is a CLI tool to convert wallpaper color schemes,
 extract color palettes, perform OCR, image compression, upscaling,
 and various image processing operations.
EOF

chmod 644 "${PKG_DIR}/DEBIAN/control"

DEB_FILE="${OUTPUT_DIR}/${PKG_NAME}-${DEB_ARCH}-linux.deb"
dpkg-deb --build --root-owner-group "${PKG_DIR}" "${DEB_FILE}"

echo "Built Debian package: ${DEB_FILE}"
rm -rf "$(dirname "${PKG_DIR}")"
