// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"testing"
)

func TestNewClient_Error(t *testing.T) {
	_, err := NewClient("", "", "", "", 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
