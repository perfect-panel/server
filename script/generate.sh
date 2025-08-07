#!/bin/bash

OS_TYPE=$(uname)
ARCH_TYPE=$(uname -m)

if [[ "$OS_TYPE" == "Linux" ]]; then
    echo "The current operating system is Linux"
    if [[ "$ARCH_TYPE" == "x86_64" ]]; then
        echo "Format api file"
        ./generate/gopure-linux-amd64 api format --dir ./apis
        echo "Architecture: amd64"
        ./generate/gopure-linux-amd64 api go -api *.api -dir . -style goZero
    elif [[ "$ARCH_TYPE" == "aarch64" ]]; then
        echo "Format api file"
        ./generate/gopure-linux-arm64 api format --dir ./apis
        echo "Architecture: arm64"
        ./generate/gopure-linux-arm64 api go -api *.api -dir . -style goZero
    else
        echo "Unrecognized architecture: $ARCH_TYPE"
    fi
elif [[ "$OS_TYPE" == "Darwin" ]]; then
    echo "The current operating system is macOS"
    if [[ "$ARCH_TYPE" == "x86_64" ]]; then
        echo "Format api file"
        ./generate/gopure-darwin-amd64 api format --dir ./apis
        echo "Architecture: amd64"
        ./generate/gopure-darwin-amd64 api go -api *.api -dir . -style goZero
    elif [[ "$ARCH_TYPE" == "arm64" ]]; then
        echo "Format api file"
        ./generate/gopure-darwin-arm64 api format --dir ./apis
        echo "Architecture: arm64"
        ./generate/gopure-darwin-arm64 api go -api *.api -dir . -style goZero
    else
        echo "Unrecognized architecture: $ARCH_TYPE"
    fi
elif [[ "$OS_TYPE" == "CYGWIN"* || "$OS_TYPE" == "MINGW"* ]]; then
    echo "The current operating system is Windows"
    if [[ "$ARCH_TYPE" == "x86_64" ]]; then
        echo "Format api file"
        ./generate/gopure-amd64.exe api format --dir ./apis
        echo "Architecture: amd64"
        ./generate/gopure-amd64.exe api go -api *.api -dir . -style goZero
    elif [[ "$ARCH_TYPE" == "arm64" ]]; then
        echo "Format api file"
        ./generate/gopure-arm64.exe api format --dir ./apis
        echo "Architecture: arm64"
        ./generate/gopure-arm64.exe api go -api *.api -dir . -style goZero
    else
        echo "Unrecognized architecture: $ARCH_TYPE"
    fi
else
    echo "Unrecognized operating system: $OS_TYPE"
fi
