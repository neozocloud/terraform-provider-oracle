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

resource "oracle_user" "test_user" {
  username = "testuser"
  password = "password"
}