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
	_ datasource.DataSource              = &permissionsDataSource{}
	_ datasource.DataSourceWithConfigure = &permissionsDataSource{}
)

func NewPermissionsDataSource() datasource.DataSource {
	return &permissionsDataSource{}
}

type permissionsDataSource struct {
	connector *queries.Connector
}

// Metadata sets the metadata for the permissions data source.
func (d *permissionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions_to_role"
}

// Schema defines the schema for the permissions data source.
func (d *permissionsDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Reads database-level permissions assigned to a role.",
		MarkdownDescription: "Reads database-level permissions assigned to a role.",
		Attributes: map[string]schema.Attribute{
			"role_name": schema.StringAttribute{
				Description:         "The database role name.",
				MarkdownDescription: "The database role name.",
				Required:            true,
			},
			"permissions": schema.ListNestedAttribute{
				Description:         "List of permissions assigned to this role.",
				MarkdownDescription: "List of permissions assigned to this role.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"permission_name": schema.StringAttribute{
							MarkdownDescription: "Permission name.",
							Computed:            true,
						},
						"class": schema.StringAttribute{
							MarkdownDescription: "Permission class.",
							Computed:            true,
						},
						"class_desc": schema.StringAttribute{
							MarkdownDescription: "Permission class description.",
							Computed:            true,
						},
						"major_id": schema.Int64Attribute{
							MarkdownDescription: "Permission Major ID.",
							Computed:            true,
						},
						"minor_id": schema.Int64Attribute{
							MarkdownDescription: "Permission Minor ID.",
							Computed:            true,
						},
						"grantee_principal_id": schema.Int64Attribute{
							MarkdownDescription: "Permission Grantee Principal ID.",
							Computed:            true,
						},
						"grantor_principal_id": schema.Int64Attribute{
							MarkdownDescription: "Permission Grantor Principal ID.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Permission type.",
							Computed:            true,
						},
						"state": schema.StringAttribute{
							MarkdownDescription: "Permission state (G=GRANT, D=DENY).",
							Computed:            true,
						},
						"state_desc": schema.StringAttribute{
							MarkdownDescription: "Permission state description.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure configures the data source with the provider configuration.
func (d *permissionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read retrieves the database permissions for a role from the database.
func (d *permissionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model.PermissionResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading database permissions for role", map[string]interface{}{
		"role_name": data.RoleName.ValueString(),
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
		Name: data.RoleName.ValueString(),
	}

	role, err = connector.GetDatabaseRole(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Database Role",
			"Could not read database role "+data.RoleName.ValueString()+": "+err.Error(),
		)
		return
	}

	// Get permissions for the role
	permissions, err := connector.GetDatabasePermissionsForRole(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Permissions",
			"Could not read permissions for role "+data.RoleName.ValueString()+": "+err.Error(),
		)
		return
	}

	// Convert to model
	permissionModels := make([]model.PermissionModel, 0, len(permissions))
	for _, perm := range permissions {
		permissionModels = append(permissionModels, model.PermissionModel{
			Class:              types.StringValue(perm.Class),
			ClassDesc:          types.StringValue(perm.ClassDesc),
			MajorID:            types.Int64Value(perm.MajorID),
			MinorID:            types.Int64Value(perm.MinorID),
			GranteePrincipalID: types.Int64Value(perm.GranteePrincipalID),
			GrantorPrincipalID: types.Int64Value(perm.GrantorPrincipalID),
			Type:               types.StringValue(perm.Type),
			Name:               types.StringValue(perm.Name),
			State:              types.StringValue(perm.State),
			StateDesc:          types.StringValue(perm.StateDesc),
		})
	}

	// Convert to types.List
	permissionsList, diags := convertPermissionsSliceToList(ctx, permissionModels)
	resp.Diagnostics.Append(*diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Permissions = permissionsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Debug(ctx, "Successfully read database permissions", map[string]interface{}{
		"role_name":         data.RoleName.ValueString(),
		"permissions_count": len(permissions),
	})
}
