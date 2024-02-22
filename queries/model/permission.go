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
