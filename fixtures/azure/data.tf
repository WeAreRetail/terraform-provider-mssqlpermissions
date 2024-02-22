data "http" "ipinfo" {
  url = "https://api.ipify.org"
}

locals {
  allowed_public_ips = toset(concat(var.allowed_ip_addresses, [data.http.ipinfo.response_body]))
}

data "azuread_group" "admin_team" {
  display_name     = var.admin_group
  security_enabled = true
}
