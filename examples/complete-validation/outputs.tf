output "created_users" {
  description = "Information about created users"
  value = {
    user_1 = {
      name         = mssqlpermissions_user.test_user_1.name
      principal_id = mssqlpermissions_user.test_user_1.principal_id
      external     = mssqlpermissions_user.test_user_1.external
    }
    user_2 = {
      name         = mssqlpermissions_user.test_user_2.name
      principal_id = mssqlpermissions_user.test_user_2.principal_id
      external     = mssqlpermissions_user.test_user_2.external
    }
  }
}

output "created_role" {
  description = "Information about created database role"
  value = {
    name             = mssqlpermissions_database_role.test_role.name
    principal_id     = mssqlpermissions_database_role.test_role.principal_id
    owning_principal = mssqlpermissions_database_role.test_role.owning_principal
    members          = mssqlpermissions_database_role_members.test_role_members.members
  }
}

output "data_source_validation" {
  description = "Validation that data sources work correctly"
  value = {
    user_data_matches = data.mssqlpermissions_user.test_user_1_data.name == mssqlpermissions_user.test_user_1.name
    role_data_matches = data.mssqlpermissions_database_role.test_role_data.name == mssqlpermissions_database_role.test_role.name
  }
}

output "permissions_granted" {
  description = "Permissions that were granted to the role"
  value       = mssqlpermissions_permissions_to_role.test_permissions.permissions
}
