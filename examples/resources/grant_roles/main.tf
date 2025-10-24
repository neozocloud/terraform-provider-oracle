terraform {
  required_providers {
    oracle = {
      source = "neozocloud/oracle"
    }
  }
}

provider "oracle" {
  host     = "localhost"
  port     = "1521"
  username = "system"
  password = "MyPassword123"
  service  = "orclpdb1"
}

resource "oracle_grant_roles" "test_grant" {
  principal   = "testuser"
  roles       = toset(["connect", "resource"])
  grants_mode = "enforce"
}