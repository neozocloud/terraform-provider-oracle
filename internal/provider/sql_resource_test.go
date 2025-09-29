// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neozocloud/terraform-provider-oracle/internal/oracle"
)

func TestAccSqlResource(t *testing.T) {
	testAccPreCheck(t)
	t.Run("basic", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: providerConfig + `
resource "oracle_sql" "test" {
  sql = "CREATE TABLE test_table (id NUMBER)"
}
`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("oracle_sql.test", "sql", "CREATE TABLE test_table (id NUMBER)"),
					),
				},
			},
		})
	})

	t.Cleanup(func() {
		dbUser := os.Getenv("ORACLE_USERNAME")
		if dbUser == "" {
			dbUser = "system"
		}
		dbPassword := os.Getenv("ORACLE_PASSWORD")
		if dbPassword == "" {
			dbPassword = "MyPassword123"
		}
		dbHost := os.Getenv("ORACLE_HOST")
		if dbHost == "" {
			dbHost = "localhost"
		}
		dbPortStr := os.Getenv("ORACLE_PORT")
		if dbPortStr == "" {
			dbPortStr = "1521"
		}
		dbServiceName := os.Getenv("ORACLE_SERVICE")
		if dbServiceName == "" {
			dbServiceName = "orclpdb1"
		}
		dbPort, err := strconv.Atoi(dbPortStr)
		if err != nil {
			t.Fatalf("failed to convert port to integer: %s", err)
		}
		client, err := oracle.NewClient(dbHost, dbServiceName, dbUser, dbPassword, dbPort)
		if err != nil {
			t.Fatalf("Failed to create client: %s", err)
		}
		_, err = client.ExecuteSQL("DROP TABLE test_table")
		if err != nil {
			t.Fatalf("Failed to drop table: %s", err)
		}
	})
}
