package model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SchemaPermissionResourceModel is the model for the schema permission resource.
// It extends the standard permission model to include schema-specific context.
type SchemaPermissionResourceModel struct {
	SchemaName  types.String `tfsdk:"schema_name"`
	RoleName    types.String `tfsdk:"role_name"`
	Permissions types.List   `tfsdk:"permissions"`
}
