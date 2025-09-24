package provider

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/provider/model"
	"terraform-provider-mssqlpermissions/internal/queries"
	qmodel "terraform-provider-mssqlpermissions/internal/queries/model"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}
var _ resource.ResourceWithConfigure = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	connector *queries.Connector
}

// Configure is called by the framework to pass provider-level configuration to the resource.
func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Metadata is a method that sets the metadata for the UserResource.
// It takes a context.Context, a resource.MetadataRequest, and a pointer to a resource.MetadataResponse as parameters.
// It sets the TypeName field of the response to the concatenation of the ProviderTypeName from the request and "_user".
func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema is a method that sets the schema for the UserResource.
// It defines the attributes and their properties for the user resource.
// The attributes include the user name, password, external flag, contained flag,
// login name, principal id, default schema, default language, object id, and SID.
func (r *UserResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "User resource.",

		Attributes: map[string]schema.Attribute{

			"name": schema.StringAttribute{
				Description:         "The user name.",
				MarkdownDescription: "The user name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Description:         "The user password.",
				MarkdownDescription: "The user password.",
				Optional:            true,
				Sensitive:           true,
			},
			"external": schema.BoolAttribute{
				Description:         "Is the user external.",
				MarkdownDescription: "Is the user external.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"principal_id": schema.Int64Attribute{
				Description:         "The user principal id.",
				MarkdownDescription: "The user principal id.",
				Computed:            true,
			},
			"default_schema": schema.StringAttribute{
				Description:         "The user default schema.",
				MarkdownDescription: "The user default schema.",
				Optional:            true,
				Computed:            true,
			},
			"default_language": schema.StringAttribute{
				Description:         "The user default language.",
				MarkdownDescription: "The user default language.",
				Optional:            true,
				Computed:            true,
			},
			"object_id": schema.StringAttribute{
				Description:         "The user object id.",
				MarkdownDescription: "The user object id.",
				Optional:            true,
			},
			"sid": schema.StringAttribute{
				Description:         "The user SID.",
				MarkdownDescription: "The user SID.",
				Computed:            true,
			},
		},
	}
}

// Create creates a new user resource in the database.
// It takes a context.Context, a resource.CreateRequest, and a pointer to a resource.CreateResponse as input parameters.
// The method retrieves the necessary information from the request, connects to the database, creates the user, and populates the state object with the created user's details.
// If any errors occur during the process, they are added to the response's diagnostics.
func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var state model.UserResourceModel
	var err error
	var diags diag.Diagnostics

	logResourceOperation(ctx, "User", "Create")

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

	user := &qmodel.User{
		Name:            state.Name.ValueString(),
		Password:        state.Password.ValueString(),
		External:        state.External.ValueBool(),
		DefaultSchema:   state.DefaultSchema.ValueString(),
		DefaultLanguage: state.DefaultLanguage.ValueString(),
		ObjectID:        state.ObjectID.ValueString(),
	}

	tflog.Debug(ctx, "Creating user")
	err = connector.CreateUser(ctx, db, user)
	if err != nil {
		resp.Diagnostics.AddError("Error creating user", err.Error())
		return
	}

	tflog.Debug(ctx, "Retrieving created user")
	user, err = connector.GetUser(ctx, db, user)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving the created user", err.Error())
		return
	}

	tflog.Debug(ctx, "Populating user state")
	state.Name = types.StringValue(user.Name)
	state.External = types.BoolValue(user.External)
	state.PrincipalID = types.Int64Value(user.PrincipalID)
	state.DefaultSchema = types.StringValue(user.DefaultSchema)
	state.DefaultLanguage = types.StringValue(user.DefaultLanguage)
	state.SID = types.StringValue(user.SID)

	if user.ObjectID == "" {
		state.ObjectID = types.StringNull()
	} else {
		state.ObjectID = types.StringValue(user.ObjectID)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperationComplete(ctx, "User", "Create")
}

// Delete is a method of the UserResource struct that handles the deletion of a user resource.
// It takes a context.Context, a resource.DeleteRequest, and a pointer to a resource.DeleteResponse as input parameters.
// The method retrieves the necessary information from the request, connects to the database, deletes the user, and populates the state object with the deleted user's details.
// If any errors occur during the process, they are added to the response's diagnostics.
func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.UserResourceModel
	var err error

	logResourceOperation(ctx, "User", "Delete")

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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

	user := &qmodel.User{
		Name: state.Name.ValueString(),
	}

	tflog.Debug(ctx, "Deleting user")
	err = connector.DeleteUser(ctx, db, user)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting user", err.Error())
		return
	}

	logResourceOperationComplete(ctx, "User", "Delete")
}

// Read is a method of the UserResource struct that handles the reading of a user resource.
// It takes a context.Context, a resource.ReadRequest, and a pointer to a resource.ReadResponse as input parameters.
// The method retrieves the necessary information from the request, connects to the database, gets the user, and populates the state object with the user's details.
// If any errors occur during the process, they are added to the response's diagnostics.
func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.UserResourceModel
	var err error

	logResourceOperation(ctx, "User", "Read")

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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

	user := &qmodel.User{
		Name:        state.Name.ValueString(),
		PrincipalID: state.PrincipalID.ValueInt64(),
		External:    state.External.ValueBool(),
	}

	tflog.Debug(ctx, "Reading user from database")
	user, err = connector.GetUser(ctx, db, user)

	// Use the centralized error handling logic
	errorResult := HandleUserReadError(err)
	if errorResult.ShouldRemoveFromState {
		tflog.Debug(ctx, "User not found in database, removing from state")
		resp.State.RemoveResource(ctx)
		return
	}

	if errorResult.ShouldAddError {
		resp.Diagnostics.AddError(errorResult.ErrorMessage, err.Error())
		return
	}

	tflog.Debug(ctx, "Populating user state")

	state.Name = types.StringValue(user.Name)
	state.External = types.BoolValue(user.External)
	state.PrincipalID = types.Int64Value(user.PrincipalID)
	state.DefaultSchema = types.StringValue(user.DefaultSchema)
	state.DefaultLanguage = types.StringValue(user.DefaultLanguage)
	state.SID = types.StringValue(user.SID)

	if user.ObjectID == "" {
		state.ObjectID = types.StringNull()
	} else {
		state.ObjectID = types.StringValue(user.ObjectID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperationComplete(ctx, "User", "Read")
}

// Update is a method of the UserResource struct that handles the updating of a user resource.
// It takes a context.Context, a resource.UpdateRequest, and a pointer to a resource.UpdateResponse as input parameters.
// The method retrieves the necessary information from the request, connects to the database, updates the user, and populates the state object with the updated user's details.
// If any errors occur during the process, they are added to the response's diagnostics.
func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var state model.UserResourceModel
	var err error

	logResourceOperation(ctx, "User", "Update")

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

	user := &qmodel.User{
		Name:            state.Name.ValueString(),
		Password:        state.Password.ValueString(),
		External:        state.External.ValueBool(),
		DefaultSchema:   state.DefaultSchema.ValueString(),
		DefaultLanguage: state.DefaultLanguage.ValueString(),
	}

	tflog.Debug(ctx, "Updating user")
	err = connector.UpdateUser(ctx, db, user)
	if err != nil {
		resp.Diagnostics.AddError("Error updating user", err.Error())
		return
	}

	tflog.Debug(ctx, "Retrieving updated user")
	user, err = connector.GetUser(ctx, db, user)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving the updated user", err.Error())
		return
	}

	tflog.Debug(ctx, "Populating updated user state")
	state.Name = types.StringValue(user.Name)
	state.External = types.BoolValue(user.External)
	state.PrincipalID = types.Int64Value(user.PrincipalID)
	state.DefaultSchema = types.StringValue(user.DefaultSchema)
	state.DefaultLanguage = types.StringValue(user.DefaultLanguage)
	state.SID = types.StringValue(user.SID)

	if user.ObjectID == "" {
		state.ObjectID = types.StringNull()
	} else {
		state.ObjectID = types.StringValue(user.ObjectID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperationComplete(ctx, "User", "Update")
}

// ImportState implements resource.ResourceWithImportState.
func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	panic("not implemented")
}
