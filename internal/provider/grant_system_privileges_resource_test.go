// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_GrantSystemPrivilegesResource(t *testing.T) {
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "oracle_user" "test_user" {
  username = "testuser_%s"
  password = "password"
}

resource "oracle_grant_system_privileges" "test_grant" {
  principal  = oracle_user.test_user.username
  privileges = toset(["CREATE SESSION"])
}
`, randString),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("oracle_grant_system_privileges.test_grant", "principal", fmt.Sprintf("testuser_%s", randString)),
					resource.TestCheckResourceAttr("oracle_grant_system_privileges.test_grant", "privileges.#", "1"),
					resource.TestCheckResourceAttr("oracle_grant_system_privileges.test_grant", "privileges.0", "CREATE SESSION"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "oracle_grant_system_privileges.test_grant",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "oracle_user" "test_user" {
  username = "testuser_%s"
  password = "password"
}

resource "oracle_grant_system_privileges" "test_grant" {
  principal  = oracle_user.test_user.username
  privileges = ["CREATE SESSION", "CREATE TABLE"]
}
`, randString),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("oracle_grant_system_privileges.test_grant", "privileges.#", "2"),
					resource.TestCheckResourceAttr("oracle_grant_system_privileges.test_grant", "privileges.0", "CREATE SESSION"),
					resource.TestCheckResourceAttr("oracle_grant_system_privileges.test_grant", "privileges.1", "CREATE TABLE"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
