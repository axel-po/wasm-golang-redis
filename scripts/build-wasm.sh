#!/usr/bin/env bash
# Compile le moteur Go en WebAssembly et place les artefacts là où le
# frontend les attend :
#   - web/public/wasmredis.wasm    le moteur (fetché par le worker)
#   - web/src/worker/wasm_exec.js  le runtime JS officiel de Go
set -euo pipefail
cd "$(dirname "$0")/.."

echo "Compilation Go -> WASM..."
GOOS=js GOARCH=wasm go build -o web/public/wasmredis.wasm ./cmd/wasm

WASM_EXEC="$(go env GOROOT)/lib/wasm/wasm_exec.js"
if [ ! -f "$WASM_EXEC" ]; then
	WASM_EXEC="$(go env GOROOT)/misc/wasm/wasm_exec.js"
fi
cp "$WASM_EXEC" web/src/worker/wasm_exec.js

echo "OK : $(du -h web/public/wasmredis.wasm | cut -f1) -> web/public/wasmredis.wasm"
