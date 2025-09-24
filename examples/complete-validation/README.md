# Complete Provider Validation Example

This example demonstrates all provider features and serves as a comprehensive validation test that you can run manually with real Terraform configurations.

## What this example tests

- ✅ **User Management**: Creates and manages SQL users (both SQL and Azure AD)
- ✅ **Database Roles**: Creates custom database roles with members
- ✅ **Role Membership**: Manages role members dynamically
- ✅ **Permissions**: Grants permissions to roles
- ✅ **Data Sources**: Validates that data sources retrieve correct information
- ✅ **Provider Configuration**: Tests both SQL and Azure AD authentication

## Usage

### Local Testing (with Docker)

1. **Start local infrastructure**:

   ```bash
   task infra:local:ready
   ```

2. **Copy and customize local configuration**:

   ```bash
   cp terraform.tfvars.local.example terraform.tfvars
   # Edit terraform.tfvars with your local settings
   ```

3. **Install the provider locally**:

   ```bash
   task install
   ```

4. **Run the validation**:

   ```bash
   terraform init
   terraform plan
   terraform apply

   # Validate outputs
   terraform output

   # Test updates (modify variables and re-apply)
   terraform plan
   terraform apply

   # Clean up
   terraform destroy
   ```

### Azure Testing

1. **Set up Azure environment**:

   ```bash
   # Source your Azure environment
   source .azure.env
   ```

2. **Copy and customize Azure configuration**:

   ```bash
   cp terraform.tfvars.azure.example terraform.tfvars
   # Edit terraform.tfvars with your Azure settings
   ```

3. **Run the validation**:

   ```bash
   terraform init
   terraform plan
   terraform apply
   terraform output
   terraform destroy
   ```

## Validation Checklist

After running this example, verify:

- [ ] All resources are created successfully
- [ ] Users are created with correct properties (internal/external)
- [ ] Database role is created with specified members
- [ ] Role members are managed correctly
- [ ] Permissions are granted to the role
- [ ] Data sources return correct information (check `data_source_validation` output)
- [ ] Resources can be updated (change members, permissions, etc.)
- [ ] Resources are destroyed cleanly

## Expected Outputs

The example produces outputs that help validate functionality:

```hcl
created_users = {
  user_1 = {
    external = false
    name = "test_user_validation_1"
    principal_id = "123"
  }
  user_2 = {
    external = false
    name = "test_user_validation_2"
    principal_id = "124"
  }
}

created_role = {
  members = ["test_user_validation_1", "test_user_validation_2"]
  name = "test_role_validation"
  owning_principal = "1"
  principal_id = "125"
}

data_source_validation = {
  role_data_matches = true
  user_data_matches = true
}

permissions_granted = [
  {
    object_name = "sys.tables"
    object_type = "OBJECT"
    permission = "SELECT"
    state = "GRANT"
  },
  {
    object_name = ""
    object_type = "DATABASE"
    permission = "CONNECT"
    state = "GRANT"
  }
]
```

## Troubleshooting

- **Provider not found**: Run `task install` to install the provider locally
- **Authentication errors**: Verify your credentials and environment variables
- **Database connection issues**: Ensure the database is accessible and credentials are correct
- **Permission errors**: Make sure the authentication principal has sufficient database permissions
