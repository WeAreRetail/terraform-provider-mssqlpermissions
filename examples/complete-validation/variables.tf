variable "server_fqdn" {
  description = "MSSQL Server FQDN"
  type        = string
}

variable "server_port" {
  description = "MSSQL Server Port"
  type        = number
  default     = 1433
}

variable "database_name" {
  description = "Database name"
  type        = string
  default     = "ApplicationDB"
}

variable "use_sql_auth" {
  description = "Use SQL authentication (for local testing)"
  type        = bool
  default     = true
}

variable "use_azure_auth" {
  description = "Use Azure AD authentication (for cloud testing)"
  type        = bool
  default     = false
}

variable "create_azure_ad_user" {
  description = "Create Azure AD users instead of SQL users for testing"
  type        = bool
  default     = false
}

variable "sql_username" {
  description = "SQL Server username (when using SQL auth)"
  type        = string
  default     = "sa"
}

variable "sql_password" {
  description = "SQL Server password (when using SQL auth)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "test_user_password" {
  description = "Password for test users (when not using Azure auth)"
  type        = string
  sensitive   = true
  default     = "UntrustedP@ssw0rd123"
}

variable "azure_tenant_id" {
  description = "Azure Tenant ID (when using Azure auth)"
  type        = string
  default     = ""
}

variable "azure_client_id" {
  description = "Azure Client ID (when using Azure auth)"
  type        = string
  default     = ""
}

variable "azure_client_secret" {
  description = "Azure Client Secret (when using Azure auth)"
  type        = string
  sensitive   = true
  default     = ""
}
