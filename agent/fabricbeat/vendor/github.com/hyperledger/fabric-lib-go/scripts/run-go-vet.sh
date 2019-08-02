#!/bin/bash -eu

#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

vet_print_funcs="Wrap,Wrapf,WithMessage"
go vet -all -printfuncs "${vet_print_funcs}" ./...
