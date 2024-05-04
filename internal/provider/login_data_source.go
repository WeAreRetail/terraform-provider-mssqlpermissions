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
	_ datasource.DataSource = &loginDataSource{}
)

func NewLoginDataSource() datasource.DataSource {
	return &loginDataSource{}
}

type loginDataSource struct {
	connector *queries.Connector
}

// Metadata is a method that sets the metadata for the loginDataSource.
// It takes a context.Context, a datasource.MetadataRequest, and a pointer to a datasource.MetadataResponse as parameters.
// It sets the TypeName field of the response to the concatenation of the ProviderTypeName from the request and "_login".
func (d *loginDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_login"
}

// Schema defines the schema for the login data source.
//
// The schema includes attributes such as the login name, ID, type, disabled status, external status,
// default database, and default language.
//
// The configuration block is shared by all resources and data sources.
func (d *loginDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Login data source.",

		Attributes: map[string]schema.Attribute{
			"config": getConfigSchema(), // config is the configuration block shared by all resources and data sources.
			"name": schema.StringAttribute{
				Description:         "Login name.",
				MarkdownDescription: "Login name.",
				Optional:            true,
				Computed:            true,
			},
			"id": schema.Int64Attribute{
				Description:         "The login ID.",
				MarkdownDescription: "The login ID.",
				Optional:            true,
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
				Computed:            true,
			},
			"default_database": schema.StringAttribute{
				Description:         "The login default database.",
				MarkdownDescription: "The login default database.",
				Computed:            true,
			},
			"default_language": schema.StringAttribute{
				Description:         "The login default language.",
				MarkdownDescription: "The login default language.",
				Computed:            true,
			},
		},
	}
}

// Read reads the login data from the database and populates the state object with the retrieved information.
// It connects to the database, retrieves the login information, and sets the values in the state object.
// If any errors occur during the process, they are added to the response diagnostics.
func (d *loginDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state model.LoginDataModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "loginDataSource: getConnector")
	d.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "loginDataSource: connect to the database")
	db, err := d.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	login := &qmodel.Login{
		Name:        state.Name.ValueString(),
		PrincipalID: state.ID.ValueInt64(),
	}

	tflog.Debug(ctx, "loginDataSource: get the login")
	login, err = d.connector.GetLogin(dbCtx, db, login)

	if err != nil {
		resp.Diagnostics.AddError("Error getting login", err.Error())
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
