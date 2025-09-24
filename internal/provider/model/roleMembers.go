package model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RoleMembersModel is the model for the role resource.
// It contains the necessary fields to configure the role.
type RoleMembersModel struct {
	Name    types.String   `tfsdk:"name"`
	Members []types.String `tfsdk:"members"`
}
