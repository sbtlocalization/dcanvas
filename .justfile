# SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: BlueOak-1.0.0

# Command runner recipes for the dcanvas project. Run `just` to list them.

# Name and output path of the CLI binary.
binary := "dcanvas"
cmd := "./cmd/dcanvas"

# List available recipes (default when `just` is run with no arguments).
default:
    @just --list

# Build the CLI binary into the repository root.
build:
    go build -o {{binary}} {{cmd}}

# Run the CLI without building a binary, e.g. `just run layout in.d.canvas`.
run *args:
    go run {{cmd}} {{args}}

# Install the CLI into $GOPATH/bin (so `dcanvas` is on PATH).
install:
    go install {{cmd}}

# Run the whole test suite. The repo is two Go modules (ADR-0012); a root
# ./... does not descend into the nested CLI module, so each is run explicitly.
test:
    go test -C . ./...
    go test -C cmd/dcanvas ./...

# Run the tests verbosely.
test-v:
    go test -C . -v ./...
    go test -C cmd/dcanvas -v ./...

# Regenerate the CLI golden test fixture.
update-golden:
    go test -C cmd/dcanvas ./ -run TestLayoutCLI_Golden -update

# Vet all packages across both modules.
vet:
    go vet -C . ./...
    go vet -C cmd/dcanvas ./...

# Format all Go source in place.
fmt:
    gofmt -w .

# Everything CI runs: vet then test.
ci: vet test

# Remove the built binary.
clean:
    rm -f {{binary}}
