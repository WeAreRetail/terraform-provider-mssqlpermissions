package provider

import (
	"context"
	"queries"
	qmodel "queries/model"
	"terraform-provider-mssqlpermissions/internal/provider/model"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &PermissionsResource{}
var _ resource.ResourceWithImportState = &PermissionsResource{}

func NewPermissionsResource() resource.Resource {
	return &PermissionsResource{}
}

type PermissionsResource struct {
	connector *queries.Connector
}

// Metadata is a method that sets the metadata for the PermissionsResource.
// It takes a context.Context, a resource.MetadataRequest, and a pointer to a resource.MetadataResponse as parameters.
// It sets the TypeName of the response to the concatenation of the ProviderTypeName from the request and "_permissions_to_role".
func (r *PermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions_to_role"
}

// Schema is a method that sets the schema for the PermissionsResource.
// It defines the attributes and their properties for the resource.
func (r *PermissionsResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Permissions.",
		MarkdownDescription: "Permissions.",
		Attributes: map[string]schema.Attribute{
			"config": getConfigSchema(), // config is the configuration block shared by all resources and data sources.

			"permissions": schema.ListNestedAttribute{
				Description:         "A list of permissions.",
				MarkdownDescription: "A list of permissions.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{

						"permission_name": schema.StringAttribute{
							MarkdownDescription: "Permission name.",
							Required:            true,
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
							MarkdownDescription: "Permission state.",
							Computed:            true,
						},

						"state_desc": schema.StringAttribute{
							MarkdownDescription: "Permission state description.",
							Computed:            true,
						},
					},
				},
			},

			"role_name": schema.StringAttribute{
				Description:         "The database role's name.",
				MarkdownDescription: "The database role's name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"is_server_role": schema.BoolAttribute{
				Description:         "Is the role a server role.",
				MarkdownDescription: "Is the role a server role.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

// Create is a method that creates a new resource based on the provided request.
// It connects to the database, confirms the role existence, and assigns permissions to the role.
// It updates the state with the newly created permissions.
// If any error occurs during the process, it adds the error to the response diagnostics.
func (r *PermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state model.PermissionResourceModel
	var err error

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "PermissionsResource: getConnector")
	r.connector = getConnector(state.Config)

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "PermissionsResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	// Confirm that the role exists.
	tflog.Debug(ctx, "PermissionsResource: confirm that the role exists")
	role := &qmodel.Role{
		Name: state.RoleName.ValueString(),
	}

	if state.IsServerRole.ValueBool() {
		role, err = r.connector.GetServerRole(dbCtx, db, role)
	} else {
		role, err = r.connector.GetDatabaseRole(dbCtx, db, role)
	}

	if err != nil {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	// Assign permissions to the role.
	tflog.Debug(ctx, "PermissionsResource: assign permissions to the role")
	var updatedPermissions []model.PermissionModel

	for _, permissionState := range state.Permissions {
		permission := &qmodel.Permission{
			Name: permissionState.Name.ValueString(),
		}

		err = r.connector.GrantPermissionToRole(dbCtx, db, role, permission)
		if err != nil {
			resp.Diagnostics.AddError("Error granting permission to role", err.Error())
			return
		}

		if state.IsServerRole.ValueBool() {
			permission, err = r.connector.GetServerPermissionForRole(dbCtx, db, role, permission)
		} else {
			permission, err = r.connector.GetDatabasePermissionForRole(dbCtx, db, role, permission)
		}

		if err != nil {
			resp.Diagnostics.AddError("Error getting permission for role", err.Error())
			return
		}

		updatedPermissions = append(
			updatedPermissions,
			model.PermissionModel{
				Class:              types.StringValue(permission.Class),
				ClassDesc:          types.StringValue(permission.ClassDesc),
				MajorID:            types.Int64Value(permission.MajorID),
				MinorID:            types.Int64Value(permission.MinorID),
				GranteePrincipalID: types.Int64Value(permission.GranteePrincipalID),
				GrantorPrincipalID: types.Int64Value(permission.GrantorPrincipalID),
				Type:               types.StringValue(permission.Type),
				Name:               types.StringValue(permission.Name),
				State:              types.StringValue(permission.State),
				StateDesc:          types.StringValue(permission.StateDesc),
			},
		)
	}

	state.Permissions = updatedPermissions

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the permissions resource.
// It removes the permissions associated with the role from the database.
// If any error occurs during the deletion process, it adds an error to the response diagnostics.
func (r *PermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.PermissionResourceModel
	var err error

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "PermissionsResource: getConnector")
	r.connector = getConnector(state.Config)

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "PermissionsResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	// Confirm that the role exists.
	tflog.Debug(ctx, "PermissionsResource: confirm that the role exists")
	role := &qmodel.Role{
		Name: state.RoleName.ValueString(),
	}

	if state.IsServerRole.ValueBool() {
		role, err = r.connector.GetServerRole(dbCtx, db, role)
	} else {
		role, err = r.connector.GetDatabaseRole(dbCtx, db, role)
	}

	if err != nil {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	// Remove permissions from the role.
	tflog.Debug(ctx, "PermissionsResource: remove permissions from the role")
	for _, permissionState := range state.Permissions {
		permission := &qmodel.Permission{
			Name: permissionState.Name.ValueString(),
		}

		err = r.connector.RevokePermissionFromRole(dbCtx, db, role, permission)
		if err != nil {
			resp.Diagnostics.AddError("Error revoking permission from role", err.Error())
			return
		}
	}

	state.Permissions = []model.PermissionModel{}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read is a method of the PermissionsResource struct that implements the resource.ReadHandler interface.
// It retrieves the state of the resource, connects to the database, confirms the existence of a role,
// retrieves the permissions for the role, and updates the state with the retrieved permissions.
// If any errors occur during the process, they are added to the response diagnostics.
func (r *PermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.PermissionResourceModel
	var err error

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "PermissionsResource: getConnector")
	r.connector = getConnector(state.Config)

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "PermissionsResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	// Confirm that the role exists.
	tflog.Debug(ctx, "PermissionsResource: confirm that the role exists")
	role := &qmodel.Role{
		Name: state.RoleName.ValueString(),
	}

	if state.IsServerRole.ValueBool() {
		role, err = r.connector.GetServerRole(dbCtx, db, role)
	} else {
		role, err = r.connector.GetDatabaseRole(dbCtx, db, role)
	}

	if err != nil {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	// Get the permissions for the role.
	tflog.Debug(ctx, "PermissionsResource: get the permissions for the role")
	var readPermissions []model.PermissionModel

	for _, permissionState := range state.Permissions {
		permission := &qmodel.Permission{
			Name: permissionState.Name.ValueString(),
		}

		if state.IsServerRole.ValueBool() {
			permission, err = r.connector.GetServerPermissionForRole(dbCtx, db, role, permission)
		} else {
			permission, err = r.connector.GetDatabasePermissionForRole(dbCtx, db, role, permission)
		}

		if err != nil && err.Error() != "permissions not found" {
			resp.Diagnostics.AddError("Error getting permission for role", err.Error())
			return
		}

		// if the permission is not found, skip it.
		if permission == nil {
			continue
		}

		readPermissions = append(
			readPermissions,
			model.PermissionModel{
				Class:              types.StringValue(permission.Class),
				ClassDesc:          types.StringValue(permission.ClassDesc),
				MajorID:            types.Int64Value(permission.MajorID),
				MinorID:            types.Int64Value(permission.MinorID),
				GranteePrincipalID: types.Int64Value(permission.GranteePrincipalID),
				GrantorPrincipalID: types.Int64Value(permission.GrantorPrincipalID),
				Type:               types.StringValue(permission.Type),
				Name:               types.StringValue(permission.Name),
				State:              types.StringValue(permission.State),
				StateDesc:          types.StringValue(permission.StateDesc),
			},
		)
	}

	state.Permissions = readPermissions

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the PermissionsResource based on the provided UpdateRequest.
// It connects to the database, confirms the existence of the role, revokes all permissions from the role,
// and then grants the updated permissions to the role.
// Finally, it updates the state of the PermissionsResource and returns any diagnostics encountered during the process.
func (r *PermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state model.PermissionResourceModel
	var plan model.PermissionResourceModel
	var err error

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "PermissionsResource: getConnector")
	r.connector = getConnector(state.Config)

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "PermissionsResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	// Confirm that the role exists.
	tflog.Debug(ctx, "PermissionsResource: confirm that the role exists")
	role := &qmodel.Role{
		Name: state.RoleName.ValueString(),
	}

	if state.IsServerRole.ValueBool() {
		role, err = r.connector.GetServerRole(dbCtx, db, role)
	} else {
		role, err = r.connector.GetDatabaseRole(dbCtx, db, role)
	}

	if err != nil {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	// As the permissions are defined in a list, the order is not guaranteed.
	// Therefore, we need to delete all permissions and then re-add them.
	// We take all the permissions in the current state and remove them.
	tflog.Debug(ctx, "PermissionsResource: delete all permissions")
	for _, permissionState := range state.Permissions {
		permission := &qmodel.Permission{
			Name: permissionState.Name.ValueString(),
		}

		err = r.connector.RevokePermissionFromRole(dbCtx, db, role, permission)
		if err != nil {
			resp.Diagnostics.AddError("Error revoking permission from role", err.Error())
			return
		}
	}

	// Now that it is clean, add the permissions.
	// We take all the permissions in the plan and add them.
	var updatedPermissions []model.PermissionModel
	for _, permissionPlan := range plan.Permissions {
		permission := &qmodel.Permission{
			Name: permissionPlan.Name.ValueString(),
		}

		err = r.connector.GrantPermissionToRole(dbCtx, db, role, permission)
		if err != nil {
			resp.Diagnostics.AddError("Error granting permission to role", err.Error())
			return
		}

		updatedPermissions = append(
			updatedPermissions,
			model.PermissionModel{
				Class:              types.StringValue(permission.Class),
				ClassDesc:          types.StringValue(permission.ClassDesc),
				MajorID:            types.Int64Value(permission.MajorID),
				MinorID:            types.Int64Value(permission.MinorID),
				GranteePrincipalID: types.Int64Value(permission.GranteePrincipalID),
				GrantorPrincipalID: types.Int64Value(permission.GrantorPrincipalID),
				Type:               types.StringValue(permission.Type),
				Name:               types.StringValue(permission.Name),
				State:              types.StringValue(permission.State),
				StateDesc:          types.StringValue(permission.StateDesc),
			},
		)
	}

	state.Permissions = updatedPermissions

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// ImportState implements resource.ResourceWithImportState.
func (r *PermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	panic("not implemented")
}
