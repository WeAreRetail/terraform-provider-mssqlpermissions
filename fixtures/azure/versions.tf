terraform {
  required_version = ">= 1.5.0"

  required_providers {
    azuread = {
      source  = "hashicorp/azuread"
      version = ">=2.45.0"
    }

    azurecaf = {
      source  = "aztfmod/azurecaf"
      version = ">=1.2.26"
    }

    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">=3.79.0"
    }

    http = {
      source  = "hashicorp/http"
      version = ">=3.4.0"
    }

    local = {
      source  = "hashicorp/local"
      version = ">=2.4.0"
    }

    random = {
      source  = "hashicorp/random"
      version = ">=3.5.1"
    }
  }
}

provider "azurerm" {
  features {}

  subscription_id     = var.subscription_id
  storage_use_azuread = true
}

provider "azuread" {
  tenant_id = var.tenant_id
}
