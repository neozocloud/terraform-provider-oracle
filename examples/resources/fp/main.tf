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

# resource "oracle_user" "test_user" {
#   username = "DETHARTEST"
#   password = "password"
# }
#
# resource "oracle_grant_system_privileges" "test_grant" {
#   principal   = oracle_user.test_user.username
#   privileges  = toset(["CREATE SESSION", "SELECT ANY TABLE"])
#   grants_mode = "enforce"
# }
#
# resource "oracle_grant_object_privileges" "test_grant" {
#   principal  = oracle_user.test_user.username
#   object     = "LP_WEBORDER.CONFIG_TYPE"
#   privileges = toset(["UPDATE", "INSERT", "DELETE"])
# }

# resource "random_password" "this" {
#   lifecycle {
#     # Allows us to import passwords that don't follow minimum requirements
#     ignore_changes = [length, lower, special, min_special, min_lower, min_numeric, min_upper]
#   }
#   for_each    = local.rds_users
#   length      = 30
#   special     = false
#   min_lower   = 1
#   min_numeric = 1
#   min_upper   = 1
#   min_special = 0
# }

resource "oracle_user" "this" {
  for_each = local.rds_users

  username           = each.key
  password           = "password"
}

resource "oracle_grant_system_privileges" "this" {
  for_each = local.rds_users

  principal   = oracle_user.this[each.key].username
  privileges  = toset(each.value.system_privileges)
  grants_mode = each.value.grants_mode
}


resource "oracle_grant_object_privileges" "this" {
  for_each = { for grant in local.user_object_grants : grant.key => grant }

  principal  = each.value.user
  object     = each.value.object
  privileges = toset(each.value.privileges)

  depends_on = [
    oracle_user.this
  ]
}
