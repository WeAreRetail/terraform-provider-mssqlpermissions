config {
  call_module_type    = "local"
  force               = false
  disabled_by_default = false
}

rule "terraform_unused_declarations" {
  enabled = false
}
