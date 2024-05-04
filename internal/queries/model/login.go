package model

// Login is the model for the login object in the MSSQL server.
type Login struct {
	Name            string
	PrincipalID     int64
	Type            string
	Is_Disabled     bool
	External        bool
	Password        string
	DefaultDatabase string
	DefaultLanguage string
}
