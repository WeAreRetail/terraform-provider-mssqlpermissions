package provider

import (
	"context"
	"strings"
	"terraform-provider-mssqlpermissions/internal/provider/model"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &SqlPermissionsProvider{}

// SqlPermissionsProvider defines the provider implementation.
type SqlPermissionsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// SqlPermissionsProviderModel describes the provider data model.
type SqlPermissionsProviderModel struct {
	ServerFqdn     types.String `tfsdk:"server_fqdn"`
	ServerPort     types.Int64  `tfsdk:"server_port"`
	DatabaseName   types.String `tfsdk:"database_name"`
	SQLLogin       types.Object `tfsdk:"sql_login"`
	SPNLogin       types.Object `tfsdk:"spn_login"`
	MSILogin       types.Object `tfsdk:"msi_login"`
	FederatedLogin types.Object `tfsdk:"federated_login"`
}

// Metadata retrieves the metadata for the mssqlpermissions provider.
// It sets the TypeName and Version fields of the MetadataResponse.
func (p *SqlPermissionsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mssqlpermissions"
	resp.Version = p.version
}

// Schema is a method that generates the schema for the provider.
// It takes a context.Context, a provider.SchemaRequest, and a pointer to a provider.SchemaResponse as parameters.
// It populates the response with the generated schema.
func (p *SqlPermissionsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage SQL Server permissions. Locally or in Azure SQL Database.",
		MarkdownDescription: "Manage SQL Server permissions. Locally or in Azure SQL Database.",
		Attributes:          getProviderConfigSchema(),
	}
}

// Configure configures the provider with the given configuration.
// It validates the configuration, creates the database connector, and makes it available to resources and data sources.
func (p *SqlPermissionsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	// Retrieve provider data from configuration.
	var config SqlPermissionsProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required fields are not empty
	if config.ServerFqdn.IsNull() || config.ServerFqdn.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("server_fqdn"),
			"Missing Server FQDN",
			"The server_fqdn is required and cannot be empty.",
		)
	}

	if config.DatabaseName.IsNull() || config.DatabaseName.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("database_name"),
			"Missing Database Name",
			"The database_name is required and cannot be empty.",
		)
	}

	// Validate server_port range
	if !config.ServerPort.IsNull() && !config.ServerPort.IsUnknown() {
		port := config.ServerPort.ValueInt64()
		if port < 1 || port > 65535 {
			resp.Diagnostics.AddAttributeError(
				path.Root("server_port"),
				"Invalid Port Number",
				"The server_port must be between 1 and 65535.",
			)
		}
	}

	// Validate authentication method mutual exclusivity
	authMethods := 0
	authMethodNames := []string{}

	if !config.SQLLogin.IsNull() && !config.SQLLogin.IsUnknown() {
		authMethods++
		authMethodNames = append(authMethodNames, "sql_login")
	}
	if !config.SPNLogin.IsNull() && !config.SPNLogin.IsUnknown() {
		authMethods++
		authMethodNames = append(authMethodNames, "spn_login")
	}
	if !config.MSILogin.IsNull() && !config.MSILogin.IsUnknown() {
		authMethods++
		authMethodNames = append(authMethodNames, "msi_login")
	}
	if !config.FederatedLogin.IsNull() && !config.FederatedLogin.IsUnknown() {
		authMethods++
		authMethodNames = append(authMethodNames, "federated_login")
	}

	if authMethods == 0 {
		resp.Diagnostics.AddError(
			"Missing Authentication Method",
			"At least one authentication method must be specified (sql_login, spn_login, msi_login, or federated_login).",
		)
	}

	if authMethods > 1 {
		resp.Diagnostics.AddError(
			"Conflicting Authentication Methods",
			"Only one authentication method can be specified at a time. Found: "+strings.Join(authMethodNames, ", "),
		)
	}

	// Return early if any validation errors occurred
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert provider model to the internal config model
	configModel := &model.ConfigModel{
		ServerFqdn:     config.ServerFqdn,
		ServerPort:     config.ServerPort,
		DatabaseName:   config.DatabaseName,
		SQLLogin:       config.SQLLogin,
		SPNLogin:       config.SPNLogin,
		MSILogin:       config.MSILogin,
		FederatedLogin: config.FederatedLogin,
	}

	// Create connector from configuration
	connector, diags := getConnector(configModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make connector available to resources via ResourceData and DataSourceData
	resp.ResourceData = connector
	resp.DataSourceData = connector
}

// Resources returns a slice of functions that create resource objects.
// Each function represents a specific resource type that can be managed by this provider.
func (p *SqlPermissionsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDatabaseRoleMembersResource,
		NewDatabaseRoleResource,
		NewPermissionsResource,
		NewSchemaPermissionsResource,
		NewUserResource,
	}
}

// DataSources returns a slice of functions that create data sources for the SQL permissions provider.
// Each function in the slice should return a datasource.DataSource.
func (p *SqlPermissionsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDatabaseRoleDataSource,
		NewDatabaseRoleMembersDataSource,
		NewPermissionsDataSource,
		NewSchemaPermissionsDataSource,
		NewUserDataSource,
	}
}

// New returns a function that creates a new instance of the SqlPermissionsProvider.
// The version parameter specifies the version of the provider.
// The returned function can be called to create a new instance of the provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SqlPermissionsProvider{
			version: version,
		}
	}
}
