package model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PermissionModel is the model for the permission included in the permission resource.
// It contains the necessary fields to configure the permission.
type PermissionModel struct {
	Class              types.String `tfsdk:"class"`
	ClassDesc          types.String `tfsdk:"class_desc"`
	MajorID            types.Int64  `tfsdk:"major_id"`
	MinorID            types.Int64  `tfsdk:"minor_id"`
	GranteePrincipalID types.Int64  `tfsdk:"grantee_principal_id"`
	GrantorPrincipalID types.Int64  `tfsdk:"grantor_principal_id"`
	Type               types.String `tfsdk:"type"`
	Name               types.String `tfsdk:"permission_name"`
	State              types.String `tfsdk:"state"`
	StateDesc          types.String `tfsdk:"state_desc"`
}

// PermissionResourceModel is the model for the permission resource.
type PermissionResourceModel struct {
	Permissions types.List   `tfsdk:"permissions"`
	RoleName    types.String `tfsdk:"role_name"`
}
