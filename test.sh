# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

export TF_ACC=true && export ORACLE_HOST=localhost && export ORACLE_PORT=1521 && export ORACLE_USERNAME=system && export ORACLE_PASSWORD=MyPassword123 && export ORACLE_SERVICE=orclpdb1 && go test -v ./... -count=1
