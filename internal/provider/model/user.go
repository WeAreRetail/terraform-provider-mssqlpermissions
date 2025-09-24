package model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UserModel is the model for the user data source.
type UserDataModel struct {
	Name            types.String `tfsdk:"name"`
	External        types.Bool   `tfsdk:"external"`
	PrincipalID     types.Int64  `tfsdk:"principal_id"`
	DefaultSchema   types.String `tfsdk:"default_schema"`
	DefaultLanguage types.String `tfsdk:"default_language"`
	ObjectID        types.String `tfsdk:"object_id"`
	SID             types.String `tfsdk:"sid"`
}

// UserResourceModel is the model for the user resource.
// It contains the necessary fields to configure the user.
type UserResourceModel struct {
	Name            types.String `tfsdk:"name"`
	Password        types.String `tfsdk:"password"`
	External        types.Bool   `tfsdk:"external"`
	PrincipalID     types.Int64  `tfsdk:"principal_id"`
	DefaultSchema   types.String `tfsdk:"default_schema"`
	DefaultLanguage types.String `tfsdk:"default_language"`
	ObjectID        types.String `tfsdk:"object_id"`
	SID             types.String `tfsdk:"sid"`
}
