package provider

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/provider/model"
	"terraform-provider-mssqlpermissions/internal/queries"
	qmodel "terraform-provider-mssqlpermissions/internal/queries/model"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &databaseRoleMembersDataSource{}
	_ datasource.DataSourceWithConfigure = &databaseRoleMembersDataSource{}
)

func NewDatabaseRoleMembersDataSource() datasource.DataSource {
	return &databaseRoleMembersDataSource{}
}

type databaseRoleMembersDataSource struct {
	connector *queries.Connector
}

// Metadata sets the metadata for the database role members data source.
func (d *databaseRoleMembersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_role_members"
}

// Schema defines the schema for the database role members data source.
func (d *databaseRoleMembersDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Reads the members of a database role.",
		MarkdownDescription: "Reads the members of a database role.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The database role name.",
				MarkdownDescription: "The database role name.",
				Required:            true,
			},
			"members": schema.ListAttribute{
				Description:         "List of user names that are members of this role.",
				MarkdownDescription: "List of user names that are members of this role.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the provider configuration.
func (d *databaseRoleMembersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	connector, ok := req.ProviderData.(*queries.Connector)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *queries.Connector, got: %T. Please report this issue to the provider developers.",
		)
		return
	}

	d.connector = connector
}

// Read retrieves the database role members from the database.
func (d *databaseRoleMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model.RoleMembersModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading database role members", map[string]interface{}{
		"role_name": data.Name.ValueString(),
	})

	connector := d.connector

	// Connect to database
	db, err := connectToDatabase(ctx, connector)
	if err != nil {
		handleDatabaseConnectionError(ctx, err, &resp.Diagnostics)
		return
	}

	// Get role information
	role := &qmodel.Role{
		Name: data.Name.ValueString(),
	}

	role, err = connector.GetDatabaseRole(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Database Role",
			"Could not read database role "+data.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// Get role members
	members, err := connector.GetDatabaseRoleMembers(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Database Role Members",
			"Could not read members for role "+data.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// Convert members to string list
	memberNames := make([]string, 0, len(members))
	for _, member := range members {
		memberNames = append(memberNames, member.Name)
	}

	// Convert to types.List
	membersList, diags := types.ListValueFrom(ctx, types.StringType, memberNames)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Members = membersList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Debug(ctx, "Successfully read database role members", map[string]interface{}{
		"role_name":    data.Name.ValueString(),
		"member_count": len(memberNames),
	})
}
