// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DirectoryResource(t *testing.T) {
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "oracle_directory" "test_dir" {
  name = "testdir_%s"
  path = "/tmp"
}
`, randString),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("oracle_directory.test_dir", "name", fmt.Sprintf("testdir_%s", randString)),
					resource.TestCheckResourceAttr("oracle_directory.test_dir", "path", "/tmp"),
				),
			},
			{
				ResourceName:      "oracle_directory.test_dir",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: providerConfig + fmt.Sprintf(`
resource "oracle_directory" "test_dir" {
  name = "testdir_%s"
  path = "/tmp/new"
}
`, randString),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("oracle_directory.test_dir", "path", "/tmp/new"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
