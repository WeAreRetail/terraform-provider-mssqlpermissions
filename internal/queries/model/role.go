// SPDX-FileCopyrightText: 2024 AWARE - Altogether We Are Retailers
// SPDX-FileContributor: Cédric Ghiot <cedric@weareretail.ai>
// SPDX-License-Identifier: MIT

package model

// Role is the model for the role object in the MSSQL server.
type Role struct {
	Name            string
	PrincipalID     int64
	Type            string
	TypeDescription string
	OwningPrincipal string
	IsFixedRole     bool
}
