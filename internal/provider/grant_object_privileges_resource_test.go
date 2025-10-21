// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_GrantObjectPrivilegesResource(t *testing.T) {
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	tableName := "test_table_" + randString
	ownerName := "test_owner_" + randString
	setupTestTableForOwner(t, tableName, ownerName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "oracle_user" "test_user" {
  username = "testuser_%s"
  password = "password"
}

resource "oracle_grant_object_privileges" "test_grant" {
  principal  = oracle_user.test_user.username
  owner      = "%s"
  object     = "%s"
  privileges = toset(["SELECT"])
}
`, randString, ownerName, tableName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("oracle_grant_object_privileges.test_grant", "principal", fmt.Sprintf("testuser_%s", randString)),
					resource.TestCheckResourceAttr("oracle_grant_object_privileges.test_grant", "owner", ownerName),
					resource.TestCheckResourceAttr("oracle_grant_object_privileges.test_grant", "object", tableName),
					resource.TestCheckResourceAttr("oracle_grant_object_privileges.test_grant", "privileges.#", "1"),
					resource.TestCheckResourceAttr("oracle_grant_object_privileges.test_grant", "privileges.0", "SELECT"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "oracle_grant_object_privileges.test_grant",
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

resource "oracle_grant_object_privileges" "test_grant" {
  principal  = oracle_user.test_user.username
  owner      = "%s"
  object     = "%s"
  privileges = toset(["SELECT", "UPDATE"])
}
`, randString, ownerName, tableName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("oracle_grant_object_privileges.test_grant", "privileges.#", "2"),
					resource.TestCheckResourceAttr("oracle_grant_object_privileges.test_grant", "privileges.0", "SELECT"),
					resource.TestCheckResourceAttr("oracle_grant_object_privileges.test_grant", "privileges.1", "UPDATE"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
	t.Cleanup(func() {
		tearDownTestTableForOwner(t, tableName, ownerName)
	})
}
