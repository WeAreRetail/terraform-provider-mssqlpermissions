package model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RoleModel is the model for the role resource.
// It contains the necessary fields to configure the role.
type RoleModel struct {
	Name            types.String `tfsdk:"name"`
	PrincipalID     types.Int64  `tfsdk:"principal_id"`
	Type            types.String `tfsdk:"type"`
	TypeDescription types.String `tfsdk:"type_description"`
	OwningPrincipal types.String `tfsdk:"owning_principal"`
	IsFixedRole     types.Bool   `tfsdk:"is_fixed_role"`
}

// RoleDataSourceModel is the model for the role data source.
// It includes members since data sources are read-only and can show current membership.
type RoleDataSourceModel struct {
	Name            types.String   `tfsdk:"name"`
	Members         []types.String `tfsdk:"members"`
	PrincipalID     types.Int64    `tfsdk:"principal_id"`
	Type            types.String   `tfsdk:"type"`
	TypeDescription types.String   `tfsdk:"type_description"`
	OwningPrincipal types.String   `tfsdk:"owning_principal"`
	IsFixedRole     types.Bool     `tfsdk:"is_fixed_role"`
}
