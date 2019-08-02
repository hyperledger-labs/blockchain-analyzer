# Copyright IBM Corp All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# -------------------------------------------------------------
# This makefile defines the followng targets

#   - all (default) - runs checks and unit-tests
#   - unit-test     - runs the go-test based unit tests
#   - checks        - runs vet and go format checks

.PHONY: all checks

all: checks unit-tests

checks: license vet gotools fmtcheck

.PHONY: fmtcheck
fmtcheck:
	@echo "Checking source code formatting..."
	@scripts/check-go-formatting.sh

.PHONY: license
license:
	@echo "Checking for license headers..."
	@scripts/check-license.sh

.PHONY: vet
vet:
	@echo "Running go vet..."
	@scripts/run-go-vet.sh

.PHONY: gotools
gotools:
	@echo "Installing goimports..."
	@go get golang.org/x/tools/cmd/goimports

.PHONY: unit-tests
unit-tests:
	@echo "Running unit tests..."
	@scripts/run-unit-tests.sh