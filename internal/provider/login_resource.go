package provider

import (
	"context"
	"queries"
	qmodel "queries/model"
	"terraform-provider-mssqlpermissions/internal/provider/model"

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

var _ resource.Resource = &LoginResource{}
var _ resource.ResourceWithImportState = &LoginResource{}

func NewLoginResource() resource.Resource {
	return &LoginResource{}
}

type LoginResource struct {
	connector *queries.Connector
}

// Metadata is a method that sets the metadata for the LoginResource.
// It takes a context.Context, a resource.MetadataRequest, and a pointer to a resource.MetadataResponse as parameters.
// It sets the TypeName of the response to the concatenation of the ProviderTypeName from the request and "_login".
func (r *LoginResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_login"
}

// Schema is a method that sets the schema for the LoginResource.
// It takes a context.Context, a resource.SchemaRequest, and a pointer to a resource.SchemaResponse as parameters.
// It sets the resp.Schema field with the desired schema for the LoginResource.
func (r *LoginResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Login data source.",

		Attributes: map[string]schema.Attribute{

			"config": getConfigSchema(), // config is the configuration block shared by all resources and data sources.
			"name": schema.StringAttribute{
				Description:         "The login name.",
				MarkdownDescription: "The login name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Description:         "The login password.",
				MarkdownDescription: "The login password.",
				Optional:            true,
				Sensitive:           true,
			},
			"id": schema.Int64Attribute{
				Description:         "The login ID.",
				MarkdownDescription: "The login ID.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				Description:         "The login type.",
				MarkdownDescription: "The login type.",
				Computed:            true,
			},
			"is_disabled": schema.BoolAttribute{
				Description:         "Is the login disabled.",
				MarkdownDescription: "Is the login disabled.",
				Computed:            true,
			},
			"external": schema.BoolAttribute{
				Description:         "Is the login external.",
				MarkdownDescription: "Is the login external.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"default_database": schema.StringAttribute{
				MarkdownDescription: "The login default database.",
				Optional:            true,
				Computed:            true,
			},
			"default_language": schema.StringAttribute{
				MarkdownDescription: "The login default language.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

// Create is a method of the LoginResource struct that handles the creation of a login resource.
// It takes a context.Context, a resource.CreateRequest, and a pointer to a resource.CreateResponse as parameters.
// It retrieves the necessary information from the request, connects to the database, creates the login, and updates the state object.
// If any errors occur during the process, they are added to the response diagnostics.
func (r *LoginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var state model.LoginResourceModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "LoginResource: getConnector")
	r.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "LoginResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	login := &qmodel.Login{
		Name:            state.Name.ValueString(),
		Password:        state.Password.ValueString(),
		External:        state.External.ValueBool(),
		DefaultDatabase: state.DefaultDatabase.ValueString(),
		DefaultLanguage: state.DefaultLanguage.ValueString(),
	}

	tflog.Debug(ctx, "LoginResource: create the login")
	err = r.connector.CreateLogin(dbCtx, db, login)

	if err != nil {
		resp.Diagnostics.AddError("Error creating login", err.Error())
		return
	}

	tflog.Debug(ctx, "LoginResource: get the created login")
	login, err = r.connector.GetLogin(dbCtx, db, login)

	if err != nil {
		resp.Diagnostics.AddError("Error retrieving the created login", err.Error())
		return
	}

	tflog.Debug(ctx, "loginDataSource: populate the state object (model.LoginModel) ")
	state.Name = types.StringValue(login.Name)
	state.ID = types.Int64Value(login.PrincipalID)
	state.Type = types.StringValue(login.Type)
	state.Is_Disabled = types.BoolValue(login.Is_Disabled)
	state.External = types.BoolValue(login.External)
	state.DefaultDatabase = types.StringValue(login.DefaultDatabase)
	state.DefaultLanguage = types.StringValue(login.DefaultLanguage)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes a login resource.
// It connects to the database using the provided connector and deletes the login specified in the state.
// If there is an error connecting to the database or deleting the login, it adds an error diagnostic to the response.
func (r *LoginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.LoginResourceModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "LoginResource: getConnector")
	r.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "LoginResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	login := &qmodel.Login{
		Name:     state.Name.ValueString(),
		External: state.External.ValueBool(),
	}

	tflog.Debug(ctx, "LoginResource: create the login")
	err = r.connector.DeleteLogin(dbCtx, db, login)

	if err != nil {
		resp.Diagnostics.AddError("Error deleting login", err.Error())
		return
	}
}

// Read reads the state of the LoginResource from the underlying database and populates the response with the retrieved data.
// It connects to the database, retrieves the login information, and updates the state object accordingly.
// If any error occurs during the process, it adds the error to the response diagnostics.
func (r *LoginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.LoginResourceModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "LoginResource: getConnector")
	r.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "LoginResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	login := &qmodel.Login{
		Name:        state.Name.ValueString(),
		PrincipalID: state.ID.ValueInt64(),
		External:    state.External.ValueBool(),
	}

	tflog.Debug(ctx, "LoginResource: get the login")
	login, err = r.connector.GetLogin(dbCtx, db, login)

	if err != nil && err.Error() != "login not found" {
		resp.Diagnostics.AddError("Error getting login", err.Error())
		return
	}

	tflog.Debug(ctx, "loginDataSource: populate the state object (model.LoginModel) ")

	if login == nil {
		state = model.LoginResourceModel{}
	} else {
		state.Name = types.StringValue(login.Name)
		state.ID = types.Int64Value(login.PrincipalID)
		state.Type = types.StringValue(login.Type)
		state.Is_Disabled = types.BoolValue(login.Is_Disabled)
		state.External = types.BoolValue(login.External)
		state.DefaultDatabase = types.StringValue(login.DefaultDatabase)
		state.DefaultLanguage = types.StringValue(login.DefaultLanguage)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the login resource with the provided request.
// It connects to the database, creates or updates the login, and retrieves the updated login information.
// The updated login information is then populated into the state object.
func (r *LoginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var state model.LoginResourceModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "LoginResource: getConnector")
	r.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "LoginResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	login := &qmodel.Login{
		Name:            state.Name.ValueString(),
		Password:        state.Password.ValueString(),
		DefaultDatabase: state.DefaultDatabase.ValueString(),
		DefaultLanguage: state.DefaultLanguage.ValueString(),
	}

	tflog.Debug(ctx, "LoginResource: create the login")
	err = r.connector.UpdateLogin(dbCtx, db, login)

	if err != nil {
		resp.Diagnostics.AddError("Error creating login", err.Error())
		return
	}

	tflog.Debug(ctx, "LoginResource: get the created login")
	login, err = r.connector.GetLogin(dbCtx, db, login)

	if err != nil {
		resp.Diagnostics.AddError("Error retrieving the created login", err.Error())
		return
	}

	tflog.Debug(ctx, "loginDataSource: populate the state object (model.LoginModel) ")
	state.Name = types.StringValue(login.Name)
	state.ID = types.Int64Value(login.PrincipalID)
	state.Type = types.StringValue(login.Type)
	state.Is_Disabled = types.BoolValue(login.Is_Disabled)
	state.External = types.BoolValue(login.External)
	state.DefaultDatabase = types.StringValue(login.DefaultDatabase)
	state.DefaultLanguage = types.StringValue(login.DefaultLanguage)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *LoginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	panic("not implemented")
}
