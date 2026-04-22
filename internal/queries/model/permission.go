// SPDX-FileCopyrightText: 2024 AWARE - Altogether We Are Retailers
// SPDX-FileContributor: Cédric Ghiot <cedric@weareretail.ai>
// SPDX-License-Identifier: MIT

package model

// Permission is the model for the permission object in the MSSQL server.
type Permission struct {
	Class              string
	ClassDesc          string
	MajorID            int64
	MinorID            int64
	GranteePrincipalID int64
	GrantorPrincipalID int64
	Type               string
	Name               string
	State              string
	StateDesc          string
}
