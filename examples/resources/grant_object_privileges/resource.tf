terraform {
  required_providers {
    oracle = {
      source = "neozocloud/oracle"
    }
  }
}

resource "oracle_grant_object_privileges" "test_grant" {
  principal  = "testuser"
  object     = "test_table"
  privileges = toset(["SELECT"])
}