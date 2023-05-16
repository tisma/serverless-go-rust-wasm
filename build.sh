#!/bin/bash

set -eux
set -o pipefail

mkdir -p target
wat2wasm wasm-modules/hello.wat -o target/hello.wasm
tinygo build -o target/hello.wasm -target=wasi wasm-modules/hellogo/hellogo.go
