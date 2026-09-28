#!/usr/bin/env bash
# Optional: download Mustangproject validator JAR for PDF/ZUGFeRD checks.
# The JAR is NOT committed to this repository. Audit works offline for XML without it.
set -euo pipefail

VERSION="${MUSTANG_VERSION:-2.16.0}"
DEST="${1:-./mustang-validator.jar}"
URL="https://github.com/ZUGFeRD/mustangproject/releases/download/v${VERSION}/Mustang-CLI-${VERSION}.jar"

echo "Fetching Mustang CLI ${VERSION} -> ${DEST}"
echo "Requires network egress (not used by go test)."
curl -fsSL -o "${DEST}" "${URL}"
echo "Done. Run: zuginbox audit --input ./invoices --mustang-jar ${DEST}"
