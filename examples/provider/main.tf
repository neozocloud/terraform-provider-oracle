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
  nne      = "enabled"
  username = "system"
  password = "MyPassword123"
  service  = "orclpdb1"
}

