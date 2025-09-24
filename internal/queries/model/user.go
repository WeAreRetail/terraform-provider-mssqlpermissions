package model

// User is the model for the user object in the MSSQL server.
type User struct {
	Name            string
	Password        string
	External        bool
	PrincipalID     int64
	DefaultSchema   string
	DefaultLanguage string
	ObjectID        string // The Azure AD object ID
	SID             string // The SID stored in the database
}
