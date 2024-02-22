package model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ConfigModel represents the configuration model for the provider.
// It contains the necessary fields to configure the connection to the MSSQL server,
// including the server FQDN, server port, database name, and login credentials.
// The login credentials can be provided in different ways, such as SQL login, SPN login,
// MSI login, or federated login.
type ConfigModel struct {
	ServerFqdn     types.String         `tfsdk:"server_fqdn"`
	ServerPort     types.Int64          `tfsdk:"server_port"`
	DatabaseName   types.String         `tfsdk:"database_name"`
	SQLLogin       *SQLLoginModel       `tfsdk:"sql_login"`
	SPNLogin       *SPNLoginModel       `tfsdk:"spn_login"`
	MSILogin       *MSILoginModel       `tfsdk:"msi_login"`
	FederatedLogin *FederatedLoginModel `tfsdk:"federated_login"`
}

// SQLLoginModel represents the SQL login model for the provider.
// It contains the necessary fields to configure the SQL login credentials,
// including the username and password.
type SQLLoginModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// SPNLoginModel represents the SPN login model for the provider.
// It contains the necessary fields to configure the SPN login credentials,
// including the client ID, client secret, and tenant ID.
type SPNLoginModel struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	TenantID     types.String `tfsdk:"tenant_id"`
}

// MSILoginModel represents the MSI login model for the provider.
// It contains the necessary fields to configure the MSI login credentials,
// including the user identity, user ID, and resource ID.
type MSILoginModel struct {
	UserIdentity types.Bool   `tfsdk:"user_identity"`
	UserId       types.String `tfsdk:"user_id"`
	ResourceId   types.String `tfsdk:"resource_id"`
}

// FederatedLoginModel represents the federated login model for the provider.
type FederatedLoginModel struct{}
