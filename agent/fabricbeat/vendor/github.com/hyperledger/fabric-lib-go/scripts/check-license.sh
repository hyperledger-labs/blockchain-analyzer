#!/bin/bash -eu

#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

source "$(dirname $0)/common.sh"

license_excluded_file_patterns=(
    ".git$"
    "(^|/)vendor/"
    "\.txt$"
    "\.md$"
    "\.rst$"
    "^Gopkg.lock$"
)

candidates=$(git_last_changed_excluding "${license_excluded_file_patterns[@]}")
files_to_check=$(filter_generated_files ${candidates})

missing=$(echo ${files_to_check} | xargs ls -d 2>/dev/null | xargs grep -L "SPDX-License-Identifier" || true)
if [ -n "$missing" ]; then
    echo "The following files are missing SPDX-License-Identifier headers:"
    echo
    echo $missing
    echo
    echo "Please replace the Apache license header comment text with:"
    echo "SPDX-License-Identifier: Apache-2.0"
    exit 1
fi
