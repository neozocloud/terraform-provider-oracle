terraform {
  required_providers {
    oracle = {
      source = "neozocloud/oracle"
    }
  }
}

provider "oracle" {
  host     = "fpprod.c36c7vwnmguc.eu-central-1.rds.amazonaws.com"
  port     = "1521"
  username = "masteruser"
  password = ""
  service  = "FPPROD"
}

resource "oracle_user" "test_user" {
  username = "DETHARTEST"
  password = "password"
}

resource "oracle_grant_system_privileges" "test_grant" {
  principal   = oracle_user.test_user.username
  privileges  = toset(["CREATE SESSION", "SELECT ANY TABLE"])
  grants_mode = "enforce"
}

# resource "oracle_grant_object_privileges" "test_grant" {
#   principal  = oracle_user.test_user.username
#   object     = "dba_users"
#   privileges = toset(["SELECT"])
# }