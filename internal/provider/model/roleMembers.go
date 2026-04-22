// SPDX-FileCopyrightText: 2024 AWARE - Altogether We Are Retailers
// SPDX-FileContributor: Cédric Ghiot <cedric@weareretail.ai>
// SPDX-License-Identifier: MIT

package model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RoleMembersModel is the model for the role resource.
// It contains the necessary fields to configure the role.
type RoleMembersModel struct {
	Name    types.String `tfsdk:"name"`
	Members types.List   `tfsdk:"members"`
}
