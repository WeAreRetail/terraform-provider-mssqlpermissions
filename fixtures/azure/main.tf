locals {
  environment     = upper(var.environment)
  project_trigram = upper(var.project_trigram)

  production = local.environment == "PRD" || local.environment == "PRE"
}

module "naming" {
  source      = "WeAreRetail/naming/azurerm"
  version     = "1.0.1"
  project     = local.project_trigram
  area        = var.area
  environment = local.environment
  location    = var.region
}

module "tagging" {
  source            = "WeAreRetail/tags/azurerm"
  version           = "1.0.0"
  rgpd_personal     = false
  rgpd_confidential = false
  budget            = var.budget
  project           = local.project_trigram
  geozone           = var.region
  disaster_recovery = var.disaster_recovery
  environment       = local.environment
  repository        = var.repository
}

module "group" {
  source  = "WeAreRetail/resource-group/azurerm"
  version = "2.0.1"

  tags           = module.tagging.tags
  location       = var.region
  caf_prefixes   = module.naming.resource_group_prefixes
  description    = var.resource_group_description
  name_separator = ""
}
