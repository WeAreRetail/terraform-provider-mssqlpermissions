package provider

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/provider/model"
	"terraform-provider-mssqlpermissions/internal/queries"
	qmodel "terraform-provider-mssqlpermissions/internal/queries/model"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &PermissionsResource{}
var _ resource.ResourceWithValidateConfig = &PermissionsResource{}
var _ resource.ResourceWithImportState = &PermissionsResource{}
var _ resource.ResourceWithConfigure = &PermissionsResource{}

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
							Optional:            true,
							Default:             stringdefault.StaticString("G"),
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
		},
	}
}

// ValidateConfig is a method that validates the configuration for the PermissionsResource.
// It checks that the role name is not empty and validates permissions configuration.
// If validation fails, it adds appropriate errors to the response diagnostics.
func (r *PermissionsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config model.PermissionResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate role_name is not empty (but allow unknown values during validation)
	if !config.RoleName.IsUnknown() && (config.RoleName.IsNull() || config.RoleName.ValueString() == "") {
		resp.Diagnostics.AddAttributeError(
			path.Root("role_name"),
			"Missing Role Name",
			"The role_name is required and cannot be empty.",
		)
	}

	// Validate permissions array is not empty (skip validation if unknown - e.g., from data source)
	if !config.Permissions.IsUnknown() && (config.Permissions.IsNull() || len(config.Permissions.Elements()) == 0) {
		resp.Diagnostics.AddAttributeError(
			path.Root("permissions"),
			"Missing Permissions",
			"At least one permission must be specified.",
		)
		return // Exit early if no permissions to validate
	}

	// Skip validation if permissions are unknown (e.g., from data source)
	if config.Permissions.IsUnknown() {
		return
	}

	// Convert permissions list to slice for validation
	permissions, diags := convertPermissionsListToSlice(ctx, config.Permissions)
	if diags != nil {
		resp.Diagnostics.Append(*diags...)
		return
	}

	// Validate each permission
	for i, permission := range permissions {
		permissionPath := path.Root("permissions").AtListIndex(i)

		// Validate permission name is not empty
		if permission.Name.IsNull() || permission.Name.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				permissionPath.AtName("permission_name"),
				"Missing Permission Name",
				"The permission_name is required and cannot be empty.",
			)
		}

		// Validate permission state is G (Grant) or D (Deny)
		if !permission.State.IsNull() {
			state := permission.State.ValueString()
			if state != "G" && state != "D" {
				resp.Diagnostics.AddAttributeError(
					permissionPath.AtName("state"),
					"Invalid Permission State",
					"The permission state must be 'G' (Grant) or 'D' (Deny).",
				)
			}
		}
	}
}

// Configure configures the resource with the provider configuration.
func (r *PermissionsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*queries.Connector)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *queries.Connector, got: %T. Please report this issue to the provider developers.",
		)
		return
	}

	r.connector = providerConfig
}

// Create is a method that creates a new resource based on the provided request.
// It connects to the database, confirms the role existence, and assigns permissions to the role.
// It updates the state with the newly created permissions.
// If any error occurs during the process, it adds the error to the response diagnostics.
func (r *PermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state model.PermissionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "PermissionsResource", "Create")

	// Use provider connector
	connector := r.connector

	// Connect to database using helper function
	db, err := connectToDatabase(ctx, connector)
	if err != nil {
		handleDatabaseConnectionError(ctx, err, &resp.Diagnostics)
		return
	}

	// Confirm that the role exists.
	role := &qmodel.Role{
		Name: state.RoleName.ValueString(),
	}

	role, err = connector.GetDatabaseRole(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	// Assign permissions to the role.
	var updatedPermissions []model.PermissionModel

	// Convert permissions list to slice for processing
	permissions, diags := convertPermissionsListToSlice(ctx, state.Permissions)
	if diags != nil {
		resp.Diagnostics.Append(*diags...)
		return
	}

	for _, permissionState := range permissions {
		permission := &qmodel.Permission{
			Name:  permissionState.Name.ValueString(),
			State: permissionState.State.ValueString(),
		}

		err = connector.AssignPermissionToRole(ctx, db, role, permission)
		if err != nil {
			resp.Diagnostics.AddError("Error granting permission to role", err.Error())
			return
		}

		permission, err = connector.GetDatabasePermissionForRole(ctx, db, role, permission)
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

	// Convert back to types.List
	updatedPermissionsList, diags := convertPermissionsSliceToList(ctx, updatedPermissions)
	if diags != nil {
		resp.Diagnostics.Append(*diags...)
		return
	}
	state.Permissions = updatedPermissionsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	logResourceOperationComplete(ctx, "PermissionsResource", "Create")
}

// Delete deletes the permissions resource.
// It removes the permissions associated with the role from the database.
// If any error occurs during the deletion process, it adds an error to the response diagnostics.
func (r *PermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.PermissionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "PermissionsResource", "Delete")

	// Use provider connector
	connector := r.connector

	// Connect to database using helper function
	db, err := connectToDatabase(ctx, connector)
	if err != nil {
		handleDatabaseConnectionError(ctx, err, &resp.Diagnostics)
		return
	}

	// Confirm that the role exists.
	role := &qmodel.Role{
		Name: state.RoleName.ValueString(),
	}

	role, err = connector.GetDatabaseRole(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	// Remove permissions from the role.
	// Convert permissions list to slice for processing
	permissions, diags := convertPermissionsListToSlice(ctx, state.Permissions)
	if diags != nil {
		resp.Diagnostics.Append(*diags...)
		return
	}

	for _, permissionState := range permissions {
		permission := &qmodel.Permission{
			Name: permissionState.Name.ValueString(),
		}

		err = connector.RevokePermissionFromRole(ctx, db, role, permission)
		if err != nil {
			resp.Diagnostics.AddError("Error revoking permission from role", err.Error())
			return
		}
	}

	// Set empty permissions list
	emptyPermissionsList, diags := convertPermissionsSliceToList(ctx, []model.PermissionModel{})
	if diags != nil {
		resp.Diagnostics.Append(*diags...)
		return
	}
	state.Permissions = emptyPermissionsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	logResourceOperationComplete(ctx, "PermissionsResource", "Delete")
}

// Read is a method of the PermissionsResource struct that implements the resource.ReadHandler interface.
// It retrieves the state of the resource, connects to the database, confirms the existence of a role,
// retrieves the permissions for the role, and updates the state with the retrieved permissions.
// If any errors occur during the process, they are added to the response diagnostics.
func (r *PermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.PermissionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "PermissionsResource", "Read")

	// Use provider connector
	connector := r.connector

	// Connect to database using helper function
	db, err := connectToDatabase(ctx, connector)
	if err != nil {
		handleDatabaseConnectionError(ctx, err, &resp.Diagnostics)
		return
	}

	// Confirm that the role exists.
	role := &qmodel.Role{
		Name: state.RoleName.ValueString(),
	}
	role, err = connector.GetDatabaseRole(ctx, db, role)

	// Use the centralized error handling logic
	errorResult := HandleDatabaseRoleReadError(err)
	if errorResult.ShouldRemoveFromState {
		tflog.Debug(ctx, "Database role not found in database, removing from state")
		resp.State.RemoveResource(ctx)
		return
	}

	if errorResult.ShouldAddError {
		resp.Diagnostics.AddError(errorResult.ErrorMessage, err.Error())
		return
	}

	// Get the permissions for the role.
	var readPermissions []model.PermissionModel

	// Convert permissions list to slice for processing
	permissions, diags := convertPermissionsListToSlice(ctx, state.Permissions)
	if diags != nil {
		resp.Diagnostics.Append(*diags...)
		return
	}

	for _, permissionState := range permissions {
		permission := &qmodel.Permission{
			Name: permissionState.Name.ValueString(),
		}

		permission, err = connector.GetDatabasePermissionForRole(ctx, db, role, permission)
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

	// Convert back to types.List
	readPermissionsList, diags := convertPermissionsSliceToList(ctx, readPermissions)
	if diags != nil {
		resp.Diagnostics.Append(*diags...)
		return
	}
	state.Permissions = readPermissionsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	logResourceOperationComplete(ctx, "PermissionsResource", "Read")
}

// Update updates the PermissionsResource based on the provided UpdateRequest.
// It connects to the database, confirms the existence of the role, revokes all permissions from the role,
// and then grants the updated permissions to the role.
// Finally, it updates the state of the PermissionsResource and returns any diagnostics encountered during the process.
func (r *PermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state model.PermissionResourceModel
	var plan model.PermissionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "PermissionsResource", "Update")

	// Use provider connector
	connector := r.connector

	// Connect to database using helper function
	db, err := connectToDatabase(ctx, connector)
	if err != nil {
		handleDatabaseConnectionError(ctx, err, &resp.Diagnostics)
		return
	}

	// Confirm that the role exists.
	role := &qmodel.Role{
		Name: state.RoleName.ValueString(),
	}

	role, err = connector.GetDatabaseRole(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	// As the permissions are defined in a list, the order is not guaranteed.
	// Therefore, we need to delete all permissions and then re-add them.
	// We take all the permissions in the current state and remove them.

	// Convert state permissions list to slice for processing
	statePermissions, diags := convertPermissionsListToSlice(ctx, state.Permissions)
	if diags != nil {
		resp.Diagnostics.Append(*diags...)
		return
	}

	for _, permissionState := range statePermissions {
		permission := &qmodel.Permission{
			Name: permissionState.Name.ValueString(),
		}

		err = connector.RevokePermissionFromRole(ctx, db, role, permission)
		if err != nil {
			resp.Diagnostics.AddError("Error revoking permission from role", err.Error())
			return
		}
	}

	// Now that it is clean, add the permissions.
	// We take all the permissions in the plan and add them.
	var updatedPermissions []model.PermissionModel

	// Convert plan permissions list to slice for processing
	planPermissions, diags := convertPermissionsListToSlice(ctx, plan.Permissions)
	if diags != nil {
		resp.Diagnostics.Append(*diags...)
		return
	}

	for _, permissionPlan := range planPermissions {
		permission := &qmodel.Permission{
			Name:  permissionPlan.Name.ValueString(),
			State: permissionPlan.State.ValueString(),
		}

		err = connector.AssignPermissionToRole(ctx, db, role, permission)
		if err != nil {
			resp.Diagnostics.AddError("Error granting permission to role", err.Error())
			return
		}

		permission, err = connector.GetDatabasePermissionForRole(ctx, db, role, permission)
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

	// Convert back to types.List
	updatedPermissionsList, diags := convertPermissionsSliceToList(ctx, updatedPermissions)
	if diags != nil {
		resp.Diagnostics.Append(*diags...)
		return
	}
	state.Permissions = updatedPermissionsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	logResourceOperationComplete(ctx, "PermissionsResource", "Update")
}

// ImportState implements resource.ResourceWithImportState.
func (r *PermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	panic("not implemented")
}

// getPermissionAttrTypes returns the attribute types for the permission model
func getPermissionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"class":                types.StringType,
		"class_desc":           types.StringType,
		"major_id":             types.Int64Type,
		"minor_id":             types.Int64Type,
		"grantee_principal_id": types.Int64Type,
		"grantor_principal_id": types.Int64Type,
		"type":                 types.StringType,
		"permission_name":      types.StringType,
		"state":                types.StringType,
		"state_desc":           types.StringType,
	}
}

// convertPermissionsListToSlice converts a types.List to []model.PermissionModel
func convertPermissionsListToSlice(ctx context.Context, permissionsList types.List) ([]model.PermissionModel, *diag.Diagnostics) {
	var permissions []model.PermissionModel
	diags := permissionsList.ElementsAs(ctx, &permissions, false)
	if diags.HasError() {
		return nil, &diags
	}
	return permissions, nil
}

// convertPermissionsSliceToList converts []model.PermissionModel to types.List
func convertPermissionsSliceToList(ctx context.Context, permissions []model.PermissionModel) (types.List, *diag.Diagnostics) {
	permissionsList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: getPermissionAttrTypes(),
	}, permissions)

	if diags.HasError() {
		return types.ListUnknown(types.ObjectType{AttrTypes: getPermissionAttrTypes()}), &diags
	}

	return permissionsList, nil
}
