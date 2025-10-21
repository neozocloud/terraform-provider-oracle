// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	grantRolesResourceName = "oracle_grant_roles.test"
)

func TestAccGrantRolesResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"oracle": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccGrantRolesResource("test_user_roles", "test_role_1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(grantRolesResourceName, "principal", "test_user_roles"),
					resource.TestCheckResourceAttr(grantRolesResourceName, "roles.#", "1"),
					resource.TestCheckResourceAttr(grantRolesResourceName, "roles.0", "test_role_1"),
				),
			},
			{
				Config: testAccGrantRolesResource("test_user_roles", "test_role_2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(grantRolesResourceName, "principal", "test_user_roles"),
					resource.TestCheckResourceAttr(grantRolesResourceName, "roles.#", "1"),
					resource.TestCheckResourceAttr(grantRolesResourceName, "roles.0", "test_role_2"),
				),
			},
		},
	})
}

func testAccGrantRolesResource(user, role string) string {
	return fmt.Sprintf(`
resource "oracle_user" "test" {
  username = %[1]q
  password = "MyPassword123"
}

resource "oracle_role" "test" {
  name = %[2]q
}

resource "oracle_grant_roles" "test" {
  principal = oracle_user.test.username
  roles = [oracle_role.test.name]
}
`, user, role)
}
