locals {
  rds_users = {
    ADDRESS_BOOK_APP = {
      system_privileges = [ "CONNECT", "resource" ]
      grants_mode = "enforce"
    },
    CDC = {
      system_privileges = [ "CONNECT", "resource" ]
      grants_mode = "enforce"
    },
    LP_CENTRALARCHIVE_PRD = {
      system_privileges = [ "CONNECT", "resource" ]
      grants_mode = "enforce"
    },
  }

  user_object_grants = flatten([
    for user, config in local.rds_users : [
      for object_name, object_config in config.object : {
        key        = "${user}-${object_name}"
        user       = user
        object     = object_name
        privileges = object_config.privileges
      }
    ]
  ])
}