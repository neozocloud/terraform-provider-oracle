// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the Oracle client is properly configured.
	// It is also important to note the ORACLE_HOST, ORACLE_PORT, ORACLE_USER,
	// ORACLE_PASSWORD and ORACLE_SERVICE environment variables must be set.
	providerConfig = `
provider "oracle" {

}
`
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server instance.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"oracle": providerserver.NewProtocol6WithError(New("test")()),
	}
)

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ORACLE_HOST"); v == "" {
		t.Fatal("ORACLE_HOST must be set for acceptance tests")
	}
	if v := os.Getenv("ORACLE_PORT"); v == "" {
		t.Fatal("ORACLE_PORT must be set for acceptance tests")
	}
	if v := os.Getenv("ORACLE_USERNAME"); v == "" {
		t.Fatal("ORACLE_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("ORACLE_PASSWORD"); v == "" {
		t.Fatal("ORACLE_PASSWORD must be set for acceptance tests")
	}
	if v := os.Getenv("ORACLE_SERVICE"); v == "" {
		t.Fatal("ORACLE_SERVICE must be set for acceptance tests")
	}
}
