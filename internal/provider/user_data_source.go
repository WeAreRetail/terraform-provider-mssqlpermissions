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
	_ datasource.DataSource = &userDataSource{}
)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct {
	connector *queries.Connector
}

// Metadata is a method that sets the metadata for the user data source.
// It takes a context.Context, a datasource.MetadataRequest, and a pointer to a datasource.MetadataResponse as parameters.
// It sets the TypeName field of the response to the concatenation of the ProviderTypeName from the request and "_user".
// The TypeName is used by the documentation generator and the language server.
// It returns nothing.
func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema is a method that sets the schema for the user data source.
// It takes a context.Context, a datasource.SchemaRequest, and a pointer to a datasource.SchemaResponse as parameters.
// It sets the Schema field of the response to a schema.Schema.
// The schema.Schema is a map of strings to schema.Attribute.
// The schema.Attribute is a struct that contains the description, markdown description, and other information about the attribute.
// The schema.Schema is used by the documentation generator and the language server.
// It returns nothing.
func (d *userDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "User data source.",

		Attributes: map[string]schema.Attribute{

			"config": getConfigSchema(), // config is the configuration block shared by all resources and data sources.
			"name": schema.StringAttribute{
				Description:         "The user name.",
				MarkdownDescription: "The user name.",
				Optional:            true,
				Computed:            true,
			},
			"external": schema.BoolAttribute{
				Description:         "Is the user external.",
				MarkdownDescription: "Is the user external.",
				Computed:            true,
			},
			"contained": schema.BoolAttribute{
				Description:         "Is the user contained.",
				MarkdownDescription: "Is the user contained.",
				Computed:            true,
			},
			"login_name": schema.StringAttribute{
				Description:         "The user login name.",
				MarkdownDescription: "The user login name.",
				Computed:            true,
			},
			"principal_id": schema.Int64Attribute{
				Description:         "The user principal id.",
				MarkdownDescription: "The user principal id.",
				Optional:            true,
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

// Read is a method that reads the user data source.
// It takes a context.Context, a datasource.ReadRequest, and a pointer to a datasource.ReadResponse as parameters.
// It sets the State field of the response to a schema.Schema.
func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state model.UserDataModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "userDataSource: getConnector")
	d.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "userDataSource: connect to the database")
	db, err := d.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	user := &qmodel.User{
		Name:        state.Name.ValueString(),
		PrincipalID: state.PrincipalID.ValueInt64(),
	}

	tflog.Debug(ctx, "userDataSource: get the user")
	user, err = d.connector.GetUser(dbCtx, db, user)

	if err != nil {
		resp.Diagnostics.AddError("Error getting user", err.Error())
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

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
