#!/bin/bash

# Copyright Terence J. Boldt (c)2021-2024
# Use of this source code is governed by an MIT
# license that can be found in the LICENSE file.

GOOS=darwin GOARCH=arm64 go build -o binaries/macos/apple-silicon/ProDOS-Utilities
GOOS=darwin GOARCH=amd64 go build -o binaries/macos/intel/ProDOS-Utilities
GOOS=windows GOARCH=amd64 go build -o binaries/windows/intel/ProDOS-Utilities.exe
GOOS=linux GOARCH=amd64 go build -o binaries/linux/intel/ProDOS-Utilities
GOOS=linux GOARCH=arm go build -o binaries/linux/arm32/ProDOS-Utilities
GOOS=linux GOARCH=arm64 go build -o binaries/linux/arm64/ProDOS-Utilities
