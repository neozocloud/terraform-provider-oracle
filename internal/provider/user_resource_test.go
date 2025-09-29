// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_UserResource(t *testing.T) {
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
`, randString),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("oracle_user.test_user", "username", fmt.Sprintf("testuser_%s", randString)),
					resource.TestCheckResourceAttr("oracle_user.test_user", "password", "password"),
					resource.TestCheckResourceAttrSet("oracle_user.test_user", "default_tablespace"),
					resource.TestCheckResourceAttrSet("oracle_user.test_user", "default_temp_tablespace"),
					resource.TestCheckResourceAttrSet("oracle_user.test_user", "profile"),
					resource.TestCheckResourceAttrSet("oracle_user.test_user", "authentication_type"),
					resource.TestCheckResourceAttrSet("oracle_user.test_user", "state"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "oracle_user.test_user",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "oracle_user" "test_user" {
  username = "testuser_%s"
  password = "newpassword"
}
`, randString),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("oracle_user.test_user", "password", "newpassword"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
