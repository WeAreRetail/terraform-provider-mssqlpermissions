package model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// LoginDataModel is the data model for the login data source.
type LoginDataModel struct {
	Config          *ConfigModel `tfsdk:"config"`
	ID              types.Int64  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
	Is_Disabled     types.Bool   `tfsdk:"is_disabled"`
	External        types.Bool   `tfsdk:"external"`
	DefaultDatabase types.String `tfsdk:"default_database"`
	DefaultLanguage types.String `tfsdk:"default_language"`
}

// LoginResourceModel is the data model for the login resource.
type LoginResourceModel struct {
	Config          *ConfigModel `tfsdk:"config"`
	ID              types.Int64  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Password        types.String `tfsdk:"password"`
	Type            types.String `tfsdk:"type"`
	Is_Disabled     types.Bool   `tfsdk:"is_disabled"`
	External        types.Bool   `tfsdk:"external"`
	DefaultDatabase types.String `tfsdk:"default_database"`
	DefaultLanguage types.String `tfsdk:"default_language"`
}
