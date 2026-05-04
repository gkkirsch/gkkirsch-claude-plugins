#!/usr/bin/env bash
# Cross-compile director-slack into bin/<os-arch>/director-slack.
# Stripped (-s -w), CGO disabled — pure-Go static binaries.
set -euo pipefail

cd "$(dirname "$0")"

TARGETS=(
  "darwin-arm64"
  "darwin-amd64"
  "linux-amd64"
  "linux-arm64"
)

rm -rf bin
for tgt in "${TARGETS[@]}"; do
  os="${tgt%-*}"
  arch="${tgt#*-}"
  out="bin/${tgt}/director-slack"
  mkdir -p "bin/${tgt}"
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 \
    go build -trimpath -ldflags="-s -w" -o "$out" .
  size=$(du -h "$out" | cut -f1)
  echo "  built $out ($size)"
done

echo
echo "all platforms built into bin/"
