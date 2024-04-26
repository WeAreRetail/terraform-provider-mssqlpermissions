package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
type SqlPermissionsProviderModel struct{}

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
		Attributes:          map[string]schema.Attribute{},
	}
}

func (p *SqlPermissionsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	// Retrieve provider data from configuration.
	var config SqlPermissionsProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}
}

// Resources returns a slice of functions that create resource objects.
// Each function represents a specific resource type and is responsible for creating the corresponding resource object.
// The order of the functions in the slice determines the order in which the resources are created.
func (p *SqlPermissionsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewLoginResource,
		NewUserResource,
		NewDatabaseRoleResource,
		NewServerRoleResource,
		NewPermissionsResource,
	}
}

// DataSources returns a slice of functions that create data sources for the SQL permissions provider.
// Each function in the slice should return a datasource.DataSource.
func (p *SqlPermissionsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewLoginDataSource,
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
