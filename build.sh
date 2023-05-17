#!/bin/bash

set -eux
set -o pipefail

mkdir -p target
wat2wasm wasm-modules/hello.wat -o target/hello.wasm
wat2wasm wasm-modules/goenv.wat -o target/goenv.wasm
tinygo build -o target/hello.wasm -target=wasi wasm-modules/hellogo/hellogo.go
tinygo build -o target/goenv.wasm -target=wasi wasm-modules/goenv/goenv.go
