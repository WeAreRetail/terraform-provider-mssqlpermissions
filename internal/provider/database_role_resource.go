package provider

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/provider/model"
	"terraform-provider-mssqlpermissions/internal/queries"
	qmodel "terraform-provider-mssqlpermissions/internal/queries/model"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &DatabaseRoleResource{}
var _ resource.ResourceWithImportState = &DatabaseRoleResource{}
var _ resource.ResourceWithConfigure = &DatabaseRoleResource{}

func NewDatabaseRoleResource() resource.Resource {
	return &DatabaseRoleResource{}
}

type DatabaseRoleResource struct {
	connector *queries.Connector
}

// Configure is called by the framework to pass provider-level configuration to the resource.
func (r *DatabaseRoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	connector, ok := req.ProviderData.(*queries.Connector)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *queries.Connector, got something else. Please report this issue to the provider developers.",
		)
		return
	}

	r.connector = connector
}

// Metadata is a method that sets the metadata for the DatabaseRoleResource.
// It takes a context.Context, a resource.MetadataRequest, and a pointer to a resource.MetadataResponse as parameters.
// It sets the TypeName field of the MetadataResponse to the concatenation of the ProviderTypeName from the MetadataRequest and "_database_role".
func (r *DatabaseRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_role"
}

// Schema is a method that sets the schema for the DatabaseRoleResource.
// It defines the attributes and their descriptions for the resource.
func (r *DatabaseRoleResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Database Role Resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The database role's name.",
				MarkdownDescription: "The database role's name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

// Create is a method of the DatabaseRoleResource struct that creates a new database role.
// It takes a context.Context, a resource.CreateRequest, and a pointer to a resource.CreateResponse as parameters.
// It connects to the database, creates the role, retrieves the created role, and adds members to the role.
// It updates the state object with the created role information.
// If any error occurs during the process, it adds the error to the response diagnostics.
func (r *DatabaseRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state model.RoleModel

	logResourceOperation(ctx, "DatabaseRole", "Create")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use provider connector
	connector := r.connector

	// Connect to database using proper context
	db, err := connectToDatabase(ctx, connector)
	if err != nil {
		handleDatabaseConnectionError(ctx, err, &resp.Diagnostics)
		return
	}

	role := &qmodel.Role{
		Name: state.Name.ValueString(),
	}

	err = connector.CreateDatabaseRole(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error creating role", err.Error())
		return
	}

	role, err = connector.GetDatabaseRole(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving the created role", err.Error())
		return
	}

	state.Name = types.StringValue(role.Name)
	state.PrincipalID = types.Int64Value(role.PrincipalID)
	state.Type = types.StringValue(role.Type)
	state.TypeDescription = types.StringValue(role.TypeDescription)
	state.OwningPrincipal = types.StringValue(role.OwningPrincipal)
	state.IsFixedRole = types.BoolValue(role.IsFixedRole)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	logResourceOperationComplete(ctx, "DatabaseRole", "Create")
}

// Delete deletes a database role.
//
// It connects to the database using the provided connector, retrieves the role information from the state,
// and deletes the role from the database.
//
// If there is an error connecting to the database or deleting the role, it adds an error diagnostic to the response.
func (r *DatabaseRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.RoleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "DatabaseRoleResource", "Delete")

	// Use provider connector
	connector := r.connector

	// Connect to database using helper function
	db, err := connectToDatabase(ctx, connector)
	if err != nil {
		handleDatabaseConnectionError(ctx, err, &resp.Diagnostics)
		return
	}

	role := &qmodel.Role{
		Name: state.Name.ValueString(),
	}

	err = connector.DeleteDatabaseRole(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting role", err.Error())
		return
	}

	logResourceOperationComplete(ctx, "DatabaseRoleResource", "Delete")
}

// Read reads the state of the DatabaseRoleResource.
// It retrieves the role information from the database and populates the state object.
// If the role is not found, it creates an empty state object.
// It returns any diagnostics encountered during the process.
func (r *DatabaseRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.RoleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "DatabaseRoleResource", "Read")

	// Use provider connector
	connector := r.connector

	// Connect to database using helper function
	db, err := connectToDatabase(ctx, connector)
	if err != nil {
		handleDatabaseConnectionError(ctx, err, &resp.Diagnostics)
		return
	}

	role := &qmodel.Role{
		Name: state.Name.ValueString(),
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

	state.Name = types.StringValue(role.Name)
	state.PrincipalID = types.Int64Value(role.PrincipalID)
	state.Type = types.StringValue(role.Type)
	state.TypeDescription = types.StringValue(role.TypeDescription)
	state.OwningPrincipal = types.StringValue(role.OwningPrincipal)
	state.IsFixedRole = types.BoolValue(role.IsFixedRole)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	logResourceOperationComplete(ctx, "DatabaseRoleResource", "Read")
}

// Update updates the database role based on the provided update request.
// Since the name requires replacement, there are no updateable fields for this resource.
func (r *DatabaseRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state model.RoleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "DatabaseRoleResource", "Update")

	// Use provider connector
	connector := r.connector

	// Connect to database using helper function
	db, err := connectToDatabase(ctx, connector)
	if err != nil {
		handleDatabaseConnectionError(ctx, err, &resp.Diagnostics)
		return
	}

	role := &qmodel.Role{
		Name: state.Name.ValueString(),
	}

	role, err = connector.GetDatabaseRole(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	state.Name = types.StringValue(role.Name)
	state.PrincipalID = types.Int64Value(role.PrincipalID)
	state.Type = types.StringValue(role.Type)
	state.TypeDescription = types.StringValue(role.TypeDescription)
	state.OwningPrincipal = types.StringValue(role.OwningPrincipal)
	state.IsFixedRole = types.BoolValue(role.IsFixedRole)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	logResourceOperationComplete(ctx, "DatabaseRoleResource", "Update")
}

// ImportState implements resource.ResourceWithImportState.
func (r *DatabaseRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	panic("not implemented")
}
