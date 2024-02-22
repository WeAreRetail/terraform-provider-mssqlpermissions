variable "admin_group" {
  type        = string
  description = "The Azure AD group part of the SQL Server admin group."
}

variable "allowed_ip_addresses" {
  type        = list(string)
  description = "The additional IP addresses to allow. The current Public IP is always granted."
  default     = []
}

variable "area" {
  type        = string
  description = "The area for naming. If unsure, use your Git branch name."
  default     = "master"
}

variable "budget" {
  type        = string
  description = "Assign a budget tag."
  default     = "TERRAFORM_PROVIDER"
}

variable "disaster_recovery" {
  type        = bool
  description = "Is this infrastructure a DR infrastrusture?"
  default     = false
}

variable "environment" {
  type        = string
  description = "The environment (DEV, INT, TST, QA, PRE, PRD)."
  default     = "DEV"
}

variable "project_trigram" {
  type        = string
  description = "The project trigram for naming convention."
  default     = "ABC"
}

variable "region" {
  type        = string
  description = "The Azure region."
  default     = "West Europe"
}

variable "repository" {
  type        = string
  description = "Assign a repository tag."
  default     = "github.com/account/repo"
}

variable "resource_group_description" {
  type        = string
  description = "The resource group description"
  default     = "Terraform provider fixtures"
}

variable "subscription_id" {
  type        = string
  description = "The subscription where to provision the infrastructure."
}

variable "tenant_id" {
  type        = string
  description = "The tenant id."
}
