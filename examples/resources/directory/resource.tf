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

resource "oracle_directory" "test_dir" {
  name = "testdir"
  path = "/tmp"
}