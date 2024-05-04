package provider

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/provider/model"
	"terraform-provider-mssqlpermissions/internal/queries"
	qmodel "terraform-provider-mssqlpermissions/internal/queries/model"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource = &databaseRoleDataSource{}
)

func NewDatabaseRoleDataSource() datasource.DataSource {
	return &databaseRoleDataSource{}
}

type databaseRoleDataSource struct {
	connector *queries.Connector
}

// Metadata is a method that sets the metadata for the user data source.
// It takes a context.Context, a datasource.MetadataRequest, and a pointer to a datasource.MetadataResponse as parameters.
// It sets the TypeName field of the response to the concatenation of the ProviderTypeName from the request and "_database_role".
// The TypeName is used by the documentation generator and the language server.
// It returns nothing.
func (d *databaseRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_role"
}

// Schema is a method that sets the schema for the user data source.
// It takes a context.Context, a datasource.SchemaRequest, and a pointer to a datasource.SchemaResponse as parameters.
// It sets the Schema field of the response to a schema.Schema.
// The schema.Schema is a map of strings to schema.Attribute.
// The schema.Attribute is a struct that contains the description, markdown description, and other information about the attribute.
// The schema.Schema is used by the documentation generator and the language server.
// It returns nothing.
func (d *databaseRoleDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Database role data source.",

		Attributes: map[string]schema.Attribute{
			"config": getConfigSchema(), // config is the configuration block shared by all resources and data sources.

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

// Read is a method that reads the database role data source.
// It takes a context.Context, a datasource.ReadRequest, and a pointer to a datasource.ReadResponse as parameters.
// It sets the State field of the response to a schema.Schema.
func (d *databaseRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state model.RoleModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "databaseRoleDataSource: getConnector")
	d.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

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
