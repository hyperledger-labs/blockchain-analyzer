#!/bin/bash -eu

#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

read -r -a source_dirs <<< $(go list -f '{{.Dir}}' ./...)

unformatted=$(gofmt -l -s "${source_dirs[@]}")
if [ -n "${unformatted}" ]; then
    echo "The following files contain gofmt errors"
    echo -e "\t${unformatted}"
    echo "The gofmt command 'gofmt -l -s -w' must be run for these files"
    exit 1
fi

import_errors=$(goimports -l "${source_dirs[@]}")
if [ -n "${import_errors}" ]; then
    echo "The following files contain goimports errors"
    echo -e "\t${import_errors}"
    echo "The goimports command 'goimports -l -w' must be run for these files"
    exit 1
fi
