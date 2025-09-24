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
	_ datasource.DataSource              = &databaseRoleDataSource{}
	_ datasource.DataSourceWithConfigure = &databaseRoleDataSource{}
)

func NewDatabaseRoleDataSource() datasource.DataSource {
	return &databaseRoleDataSource{}
}

type databaseRoleDataSource struct {
	connector *queries.Connector
}

// Metadata sets the metadata for the database role data source.
// It sets the TypeName to include the provider type name and "_database_role".
func (d *databaseRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_role"
}

// Schema defines the schema for the database role data source.
// It specifies the available attributes that can be configured or computed.
func (d *databaseRoleDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Database role data source.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The database role's name.",
				MarkdownDescription: "The database role's name.",
				Optional:            true,
				Computed:            true,
			},
			"members": schema.ListAttribute{
				Description:         "The database role's members.",
				MarkdownDescription: "The database role's members.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"principal_id": schema.Int64Attribute{
				Description:         "Database role principal id.",
				MarkdownDescription: "Database role principal id.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				Description:         "Database role type.",
				MarkdownDescription: "Database role type.",
				Computed:            true,
			},
			"type_description": schema.StringAttribute{
				Description:         "Database role type description.",
				MarkdownDescription: "Database role type description.",
				Computed:            true,
			},
			"owning_principal": schema.StringAttribute{
				Description:         "Database role owning principal.",
				MarkdownDescription: "Database role owning principal.",
				Computed:            true,
			},
			"is_fixed_role": schema.BoolAttribute{
				Description:         "Is the database role a fixed role.",
				MarkdownDescription: "Is the database role a fixed role.",
				Computed:            true,
			},
		},
	}
}

// Configure is called by the framework to pass provider-level configuration to the data source.
func (d *databaseRoleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	connector, ok := req.ProviderData.(*queries.Connector)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			"Expected *queries.Connector, got something else. Please report this issue to the provider developers.",
		)
		return
	}

	d.connector = connector
}

// Read is a method of the databaseRoleDataSource struct that reads the state of the data source.
// It retrieves the role information from the database and populates the state object.
// If the role is not found, it creates an empty state object.
// It returns any diagnostics encountered during the process.
func (d *databaseRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state model.RoleDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "databaseRoleDataSource: using provider connector")

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "databaseRoleDataSource: connect to the database")
	db, err := d.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	role := &qmodel.Role{
		Name: state.Name.ValueString(),
	}

	tflog.Debug(ctx, "databaseRoleDataSource: get the user")
	role, err = d.connector.GetDatabaseRole(dbCtx, db, role)

	if err != nil {
		resp.Diagnostics.AddError("Error getting database role", err.Error())
		return
	}

	tflog.Debug(ctx, "databaseRoleDataSource: populate the state object (model.RoleModel) ")
	state.Name = types.StringValue(role.Name)
	state.PrincipalID = types.Int64Value(role.PrincipalID)
	state.Type = types.StringValue(role.Type)
	state.TypeDescription = types.StringValue(role.TypeDescription)
	state.OwningPrincipal = types.StringValue(role.OwningPrincipal)
	state.IsFixedRole = types.BoolValue(role.IsFixedRole)

	var members []*qmodel.User
	members, err = d.connector.GetDatabaseRoleMembers(dbCtx, db, role)

	if err != nil {
		resp.Diagnostics.AddError("Error getting database role members", err.Error())
		return
	}

	// Get all members name and add to state.members
	for _, member := range members {
		state.Members = append(state.Members, types.StringValue(member.Name))
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
