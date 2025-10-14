package provider

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/provider/model"
	"terraform-provider-mssqlpermissions/internal/queries"
	qmodel "terraform-provider-mssqlpermissions/internal/queries/model"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &SchemaPermissionsResource{}
var _ resource.ResourceWithValidateConfig = &SchemaPermissionsResource{}
var _ resource.ResourceWithImportState = &SchemaPermissionsResource{}
var _ resource.ResourceWithConfigure = &SchemaPermissionsResource{}

func NewSchemaPermissionsResource() resource.Resource {
	return &SchemaPermissionsResource{}
}

type SchemaPermissionsResource struct {
	connector *queries.Connector
}

// Metadata sets the metadata for the SchemaPermissionsResource.
func (r *SchemaPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schema_permissions"
}

// Schema defines the schema for the SchemaPermissionsResource.
func (r *SchemaPermissionsResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Schema-level permissions assigned to a database role.",
		MarkdownDescription: "Schema-level permissions assigned to a database role.",
		Attributes: map[string]schema.Attribute{
			"schema_name": schema.StringAttribute{
				Description:         "The schema name.",
				MarkdownDescription: "The schema name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

			"permissions": schema.ListNestedAttribute{
				Description:         "A list of permissions on the schema.",
				MarkdownDescription: "A list of permissions on the schema.",
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
							MarkdownDescription: "Permission state (G=GRANT, D=DENY).",
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
		},
	}
}

// ValidateConfig validates the configuration for the SchemaPermissionsResource.
func (r *SchemaPermissionsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config model.SchemaPermissionResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate schema_name is not empty
	if !config.SchemaName.IsUnknown() && (config.SchemaName.IsNull() || config.SchemaName.ValueString() == "") {
		resp.Diagnostics.AddAttributeError(
			path.Root("schema_name"),
			"Missing Schema Name",
			"The schema_name is required and cannot be empty.",
		)
	}

	// Validate role_name is not empty
	if !config.RoleName.IsUnknown() && (config.RoleName.IsNull() || config.RoleName.ValueString() == "") {
		resp.Diagnostics.AddAttributeError(
			path.Root("role_name"),
			"Missing Role Name",
			"The role_name is required and cannot be empty.",
		)
	}

	// Validate permissions array is not empty
	if !config.Permissions.IsUnknown() && (config.Permissions.IsNull() || len(config.Permissions.Elements()) == 0) {
		resp.Diagnostics.AddAttributeError(
			path.Root("permissions"),
			"Missing Permissions",
			"At least one permission must be specified.",
		)
		return
	}

	// Skip validation if permissions are unknown
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
					"The permission state must be either 'G' (GRANT) or 'D' (DENY).",
				)
			}
		}
	}
}

// Configure configures the resource with the provider configuration.
func (r *SchemaPermissionsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new schema permissions resource.
func (r *SchemaPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state model.SchemaPermissionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "SchemaPermissionsResource", "Create")

	connector := r.connector

	// Connect to database using helper function
	db, err := connectToDatabase(ctx, connector)
	if err != nil {
		handleDatabaseConnectionError(ctx, err, &resp.Diagnostics)
		return
	}

	// Confirm that the role exists
	role := &qmodel.Role{
		Name: state.RoleName.ValueString(),
	}

	role, err = connector.GetDatabaseRole(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	schemaName := state.SchemaName.ValueString()

	// Assign permissions to the role on the schema
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

		err = connector.AssignPermissionOnSchemaToRole(ctx, db, role, schemaName, permission)
		if err != nil {
			resp.Diagnostics.AddError("Error granting permission on schema to role", err.Error())
			return
		}

		permission, err = connector.GetSchemaPermissionForRole(ctx, db, role, schemaName, permission)
		if err != nil {
			resp.Diagnostics.AddError("Error getting permission for role on schema", err.Error())
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
	logResourceOperationComplete(ctx, "SchemaPermissionsResource", "Create")
}

// Read reads the schema permissions resource.
func (r *SchemaPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.SchemaPermissionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "SchemaPermissionsResource", "Read")

	connector := r.connector

	// Connect to database using helper function
	db, err := connectToDatabase(ctx, connector)
	if err != nil {
		handleDatabaseConnectionError(ctx, err, &resp.Diagnostics)
		return
	}

	// Confirm that the role exists
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

	schemaName := state.SchemaName.ValueString()

	// Get the permissions for the role on the schema
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

		permission, err = connector.GetSchemaPermissionForRole(ctx, db, role, schemaName, permission)
		if err != nil && err.Error() != "permissions not found" {
			resp.Diagnostics.AddError("Error getting permission for role on schema", err.Error())
			return
		}

		// If the permission is not found, skip it
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
	logResourceOperationComplete(ctx, "SchemaPermissionsResource", "Read")
}

// Update updates the schema permissions resource.
func (r *SchemaPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state model.SchemaPermissionResourceModel
	var plan model.SchemaPermissionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "SchemaPermissionsResource", "Update")

	connector := r.connector

	// Connect to database using helper function
	db, err := connectToDatabase(ctx, connector)
	if err != nil {
		handleDatabaseConnectionError(ctx, err, &resp.Diagnostics)
		return
	}

	// Confirm that the role exists
	role := &qmodel.Role{
		Name: state.RoleName.ValueString(),
	}

	role, err = connector.GetDatabaseRole(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	schemaName := state.SchemaName.ValueString()

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

		err = connector.RevokePermissionOnSchemaFromRole(ctx, db, role, schemaName, permission)
		if err != nil {
			resp.Diagnostics.AddError("Error revoking permission from role on schema", err.Error())
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

		err = connector.AssignPermissionOnSchemaToRole(ctx, db, role, schemaName, permission)
		if err != nil {
			resp.Diagnostics.AddError("Error granting permission on schema to role", err.Error())
			return
		}

		permission, err = connector.GetSchemaPermissionForRole(ctx, db, role, schemaName, permission)
		if err != nil {
			resp.Diagnostics.AddError("Error getting permission for role on schema", err.Error())
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
	logResourceOperationComplete(ctx, "SchemaPermissionsResource", "Update")
}

// Delete deletes the schema permissions resource.
func (r *SchemaPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.SchemaPermissionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "SchemaPermissionsResource", "Delete")

	connector := r.connector

	// Connect to database using helper function
	db, err := connectToDatabase(ctx, connector)
	if err != nil {
		handleDatabaseConnectionError(ctx, err, &resp.Diagnostics)
		return
	}

	// Confirm that the role exists
	role := &qmodel.Role{
		Name: state.RoleName.ValueString(),
	}

	role, err = connector.GetDatabaseRole(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	schemaName := state.SchemaName.ValueString()

	// Remove permissions from the role on the schema
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

		err = connector.RevokePermissionOnSchemaFromRole(ctx, db, role, schemaName, permission)
		if err != nil {
			resp.Diagnostics.AddError("Error revoking permission from role on schema", err.Error())
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
	logResourceOperationComplete(ctx, "SchemaPermissionsResource", "Delete")
}

// ImportState implements resource.ResourceWithImportState.
func (r *SchemaPermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import is not implemented for this resource as it requires complex state reconstruction
	resp.Diagnostics.AddError(
		"Import Not Supported",
		"Importing schema permissions is not currently supported. Please define the resource in your Terraform configuration.",
	)
}
