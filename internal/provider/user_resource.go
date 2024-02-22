package provider

import (
	"context"
	"queries"
	qmodel "queries/model"
	"terraform-provider-mssqlpermissions/internal/provider/model"

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

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	connector *queries.Connector
}

// Metadata is a method that sets the metadata for the UserResource.
// It takes a context.Context, a resource.MetadataRequest, and a pointer to a resource.MetadataResponse as parameters.
// It sets the TypeName field of the response to the concatenation of the ProviderTypeName from the request and "_user".
func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema is a method that sets the schema for the UserResource.
// It defines the attributes and their properties for the user data source.
// The attributes include the user name, password, external flag, contained flag,
// login name, principal id, default schema, default language, object id, and SID.
func (r *UserResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "User data source.",

		Attributes: map[string]schema.Attribute{

			"config": getConfigSchema(), // config is the configuration block shared by all resources and data sources.

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
			"contained": schema.BoolAttribute{
				Description:         "Is the user contained.",
				MarkdownDescription: "Is the user contained.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"login_name": schema.StringAttribute{
				Description:         "The user login name.",
				MarkdownDescription: "The user login name.",
				Optional:            true,
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

// Create is a method of the UserResource struct that handles the creation of a user resource.
// It takes a context.Context, a resource.CreateRequest, and a pointer to a resource.CreateResponse as input parameters.
// The method retrieves the necessary information from the request, connects to the database, creates the user, and populates the state object with the created user's details.
// If any errors occur during the process, they are added to the response's diagnostics.
func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var state model.UserResourceModel
	var err error

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "UserResource: getConnector")
	r.connector = getConnector(state.Config)

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "UserResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	user := &qmodel.User{
		Name:            state.Name.ValueString(),
		Password:        state.Password.ValueString(),
		External:        state.External.ValueBool(),
		Contained:       state.Contained.ValueBool(),
		LoginName:       state.LoginName.ValueString(),
		DefaultSchema:   state.DefaultSchema.ValueString(),
		DefaultLanguage: state.DefaultLanguage.ValueString(),
		ObjectID:        state.ObjectID.ValueString(),
	}

	tflog.Debug(ctx, "UserResource: create the user")
	err = r.connector.CreateUser(dbCtx, db, user)

	if err != nil {
		resp.Diagnostics.AddError("Error creating user", err.Error())
		return
	}

	tflog.Debug(ctx, "UserResource: get the created user")
	user, err = r.connector.GetUser(dbCtx, db, user)

	if err != nil {
		resp.Diagnostics.AddError("Error retrieving the created user", err.Error())
		return
	}

	tflog.Debug(ctx, "userDataSource: populate the state object (model.UserModel) ")
	state.Name = types.StringValue(user.Name)
	state.External = types.BoolValue(user.External)
	state.Contained = types.BoolValue(user.Contained)
	state.PrincipalID = types.Int64Value(user.PrincipalID)
	state.DefaultSchema = types.StringValue(user.DefaultSchema)
	state.DefaultLanguage = types.StringValue(user.DefaultLanguage)
	state.SID = types.StringValue(user.SID)

	if user.LoginName == "" {
		state.LoginName = types.StringNull()
	} else {
		state.LoginName = types.StringValue(user.LoginName)
	}

	if user.ObjectID == "" {
		state.ObjectID = types.StringNull()
	} else {
		state.ObjectID = types.StringValue(user.ObjectID)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete is a method of the UserResource struct that handles the deletion of a user resource.
// It takes a context.Context, a resource.DeleteRequest, and a pointer to a resource.DeleteResponse as input parameters.
// The method retrieves the necessary information from the request, connects to the database, deletes the user, and populates the state object with the deleted user's details.
// If any errors occur during the process, they are added to the response's diagnostics.
func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.UserResourceModel
	var err error

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "UserResource: getConnector")
	r.connector = getConnector(state.Config)

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "UserResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	user := &qmodel.User{
		Name: state.Name.ValueString(),
	}

	tflog.Debug(ctx, "UserResource: create the user")
	err = r.connector.DeleteUser(dbCtx, db, user)

	if err != nil {
		resp.Diagnostics.AddError("Error deleting user", err.Error())
		return
	}
}

// Read is a method of the UserResource struct that handles the reading of a user resource.
// It takes a context.Context, a resource.ReadRequest, and a pointer to a resource.ReadResponse as input parameters.
// The method retrieves the necessary information from the request, connects to the database, gets the user, and populates the state object with the user's details.
// If any errors occur during the process, they are added to the response's diagnostics.
func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.UserResourceModel
	var err error

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "UserResource: getConnector")
	r.connector = getConnector(state.Config)

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "UserResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	user := &qmodel.User{
		Name:        state.Name.ValueString(),
		PrincipalID: state.PrincipalID.ValueInt64(),
		External:    state.External.ValueBool(),
	}

	tflog.Debug(ctx, "UserResource: get the user")
	user, err = r.connector.GetUser(dbCtx, db, user)

	if err != nil && err.Error() != "user not found" {
		resp.Diagnostics.AddError("Error getting user", err.Error())
		return
	}

	tflog.Debug(ctx, "userDataSource: populate the state object (model.UserModel) ")

	if user == nil {
		state = model.UserResourceModel{
			Config: state.Config,
		}
	} else {
		state.Name = types.StringValue(user.Name)
		state.External = types.BoolValue(user.External)
		state.Contained = types.BoolValue(user.Contained)
		state.LoginName = types.StringValue(user.LoginName)
		state.PrincipalID = types.Int64Value(user.PrincipalID)
		state.DefaultSchema = types.StringValue(user.DefaultSchema)
		state.DefaultLanguage = types.StringValue(user.DefaultLanguage)
		state.SID = types.StringValue(user.SID)
	}

	if user.LoginName == "" {
		state.LoginName = types.StringNull()
	} else {
		state.LoginName = types.StringValue(user.LoginName)
	}

	if user.ObjectID == "" {
		state.ObjectID = types.StringNull()
	} else {
		state.ObjectID = types.StringValue(user.ObjectID)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update is a method of the UserResource struct that handles the updating of a user resource.
// It takes a context.Context, a resource.UpdateRequest, and a pointer to a resource.UpdateResponse as input parameters.
// The method retrieves the necessary information from the request, connects to the database, updates the user, and populates the state object with the updated user's details.
// If any errors occur during the process, they are added to the response's diagnostics.
func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var state model.UserResourceModel
	var err error

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "UserResource: getConnector")
	r.connector = getConnector(state.Config)

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "UserResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	user := &qmodel.User{
		Name:            state.Name.ValueString(),
		Password:        state.Password.ValueString(),
		External:        state.External.ValueBool(),
		Contained:       state.Contained.ValueBool(),
		LoginName:       state.LoginName.ValueString(),
		DefaultSchema:   state.DefaultSchema.ValueString(),
		DefaultLanguage: state.DefaultLanguage.ValueString(),
	}

	tflog.Debug(ctx, "UserResource: create the user")
	err = r.connector.UpdateUser(dbCtx, db, user)

	if err != nil {
		resp.Diagnostics.AddError("Error creating user", err.Error())
		return
	}

	tflog.Debug(ctx, "UserResource: get the created user")
	user, err = r.connector.GetUser(dbCtx, db, user)

	if err != nil {
		resp.Diagnostics.AddError("Error retrieving the created user", err.Error())
		return
	}

	tflog.Debug(ctx, "userDataSource: populate the state object (model.UserModel) ")
	state.Name = types.StringValue(user.Name)
	state.External = types.BoolValue(user.External)
	state.Contained = types.BoolValue(user.Contained)
	state.PrincipalID = types.Int64Value(user.PrincipalID)
	state.DefaultSchema = types.StringValue(user.DefaultSchema)
	state.DefaultLanguage = types.StringValue(user.DefaultLanguage)
	state.SID = types.StringValue(user.SID)

	if user.LoginName == "" {
		state.LoginName = types.StringNull()
	} else {
		state.LoginName = types.StringValue(user.LoginName)
	}

	if user.ObjectID == "" {
		state.ObjectID = types.StringNull()
	} else {
		state.ObjectID = types.StringValue(user.ObjectID)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	panic("not implemented")
}
