package provider

import (
	"context"
	"terraform-provider-mssqlpermissions/internal/provider/model"
	"terraform-provider-mssqlpermissions/internal/queries"
	qmodel "terraform-provider-mssqlpermissions/internal/queries/model"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &ServerRoleResource{}
var _ resource.ResourceWithImportState = &ServerRoleResource{}

func NewServerRoleResource() resource.Resource {
	return &ServerRoleResource{}
}

type ServerRoleResource struct {
	connector *queries.Connector
}

// Metadata is a method that sets the metadata for the server role resource.
// It takes a context.Context, a resource.MetadataRequest, and a pointer to a resource.MetadataResponse as parameters.
// The method sets the TypeName field of the response to the concatenation of the ProviderTypeName from the request and "_server_role".
func (r *ServerRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_role"
}

// Schema is a method that sets the schema for the ServerRoleResource.
// It takes a context.Context, a resource.SchemaRequest, and a pointer to a resource.SchemaResponse as parameters.
// It sets the resp.Schema field with the schema for the ServerRoleResource.
// The schema includes attributes such as config, name, members, principal_id, type, type_description, owning_principal, and is_fixed_role.
// Each attribute has a description and other properties such as whether it is required or computed.
func (r *ServerRoleResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Server Role Resource.",

		Attributes: map[string]schema.Attribute{

			"config": getConfigSchema(), // config is the configuration block shared by all resources and data sources.

			"name": schema.StringAttribute{
				Description:         "The server role's name.",
				MarkdownDescription: "The server role's name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"members": schema.ListAttribute{
				Description:         "The server role's members.",
				MarkdownDescription: "The server role's members.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"principal_id": schema.Int64Attribute{
				Description:         "Server Role principal id.",
				MarkdownDescription: "Server Role principal id.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				Description:         "Server Role type.",
				MarkdownDescription: "Server Role type.",
				Computed:            true,
			},
			"type_description": schema.StringAttribute{
				Description:         "Server Role type description.",
				MarkdownDescription: "Server Role type description.",
				Computed:            true,
			},
			"owning_principal": schema.StringAttribute{
				Description:         "Server Role owning principal.",
				MarkdownDescription: "Server Role owning principal.",
				Computed:            true,
			},
			"is_fixed_role": schema.BoolAttribute{
				Description:         "Is the server role a fixed role.",
				MarkdownDescription: "Is the server role a fixed role.",
				Computed:            true,
			},
		},
	}
}

// Create is a method of the ServerRoleResource struct that creates a new server role in the database.
// It takes a context.Context, a resource.CreateRequest, and a pointer to a resource.CreateResponse as input parameters.
// It populates the response with any diagnostics or errors encountered during the creation process.
// If there are any errors, the method returns without making any changes.
// Otherwise, it connects to the database, creates the role, retrieves the created role, populates the state object,
// adds members to the role, and updates the response state with the new state object.
func (r *ServerRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var state model.RoleModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "ServerRoleResource: getConnector")
	r.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "ServerRoleResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	role := &qmodel.Role{
		Name: state.Name.ValueString(),
	}

	tflog.Debug(ctx, "ServerRoleResource: create the role")
	err = r.connector.CreateServerRole(dbCtx, db, role)

	if err != nil {
		resp.Diagnostics.AddError("Error creating role", err.Error())
		return
	}

	tflog.Debug(ctx, "ServerRoleResource: get the created role")
	role, err = r.connector.GetServerRole(dbCtx, db, role)

	if err != nil {
		resp.Diagnostics.AddError("Error retrieving the created role", err.Error())
		return
	}

	tflog.Debug(ctx, "ServerRoleResource: populate the state object (model.RoleModel) ")
	state.Name = types.StringValue(role.Name)
	state.PrincipalID = types.Int64Value(role.PrincipalID)
	state.Type = types.StringValue(role.Type)
	state.TypeDescription = types.StringValue(role.TypeDescription)
	state.OwningPrincipal = types.StringValue(role.OwningPrincipal)
	state.IsFixedRole = types.BoolValue(role.IsFixedRole)

	// Add the members to the role
	for _, member := range state.Members {
		login := &qmodel.Login{
			Name: member.ValueString(),
		}

		tflog.Debug(ctx, "ServerRoleResource: add the member to the role")
		err = r.connector.AddServerRoleMember(dbCtx, db, role, login)
		if err != nil {
			resp.Diagnostics.AddError("Error adding login to role", err.Error())
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes a server role.
// It retrieves the role state from the request, connects to the database using the connector,
// and deletes the role from the database. If any error occurs during the process, it adds
// an error diagnostic to the response.
func (r *ServerRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.RoleModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "ServerRoleResource: getConnector")
	r.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "ServerRoleResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	role := &qmodel.Role{
		Name: state.Name.ValueString(),
	}

	tflog.Debug(ctx, "ServerRoleResource: create the role")
	err = r.connector.DeleteServerRole(dbCtx, db, role)

	if err != nil {
		resp.Diagnostics.AddError("Error deleting role", err.Error())
		return
	}
}

// Read reads the server role resource from the database and populates the state object.
// It connects to the database using the connector and retrieves the role information.
// If the role is not found, it sets the state object with default values.
// The populated state object is then set in the response.
func (r *ServerRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.RoleModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "ServerRoleResource: getConnector")
	r.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "ServerRoleResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	role := &qmodel.Role{
		Name: state.Name.ValueString(),
	}

	tflog.Debug(ctx, "ServerRoleResource: get the role")
	role, err = r.connector.GetServerRole(dbCtx, db, role)

	if err != nil && err.Error() != "server role not found" {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	tflog.Debug(ctx, "ServerRoleResource: populate the state object (model.RoleModel) ")

	if role == nil {
		state = model.RoleModel{
			Config: state.Config,
		}
	} else {

		// Get the members of the role
		members, err := r.connector.GetServerRoleMembers(ctx, db, role)
		if err != nil {
			resp.Diagnostics.AddError("Error getting role members", err.Error())
			return
		}

		// Convert the members to a list of string values
		state.Members = make([]types.String, 0)
		for _, member := range members {
			state.Members = append(state.Members, types.StringValue(member.Name))
		}

		state.Name = types.StringValue(role.Name)
		state.PrincipalID = types.Int64Value(role.PrincipalID)
		state.Type = types.StringValue(role.Type)
		state.TypeDescription = types.StringValue(role.TypeDescription)
		state.OwningPrincipal = types.StringValue(role.OwningPrincipal)
		state.IsFixedRole = types.BoolValue(role.IsFixedRole)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the server role resource.
// It updates only the members of the role.
// It compares the members in the plan with the members in the database,
// adds the members that are in the plan but not in the database,
// and removes the members that are in the database but not in the plan.
// Finally, it populates the state object with the updated role information.
func (r *ServerRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state model.RoleModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "ServerRoleResource: getConnector")
	r.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "ServerRoleResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	role := &qmodel.Role{
		Name: state.Name.ValueString(),
	}

	tflog.Debug(ctx, "ServerRoleResource: get the role")
	role, err = r.connector.GetServerRole(dbCtx, db, role)
	if err != nil && err.Error() != "server role not found" {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	tflog.Debug(ctx, "ServerRoleResource: get the role members")
	membersInDB, err := r.connector.GetServerRoleMembers(dbCtx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error getting role members", err.Error())
		return
	}

	var loginsToAdd []*qmodel.Login = make([]*qmodel.Login, 0)
	var loginsToRemove []*qmodel.Login = make([]*qmodel.Login, 0)

	// Compare the members in the plan with the members in the database.
	// If the member is in the plan but not in the database, add it.
	for _, memberInPlan := range state.Members {
		found := false
		for _, memberInDB := range membersInDB {
			if memberInPlan.ValueString() == memberInDB.Name {
				found = true
			}
		}
		if !found {
			login := &qmodel.Login{
				Name: memberInPlan.ValueString(),
			}
			loginsToAdd = append(loginsToAdd, login)
		}
	}

	// Compare the members in the database with the members in the plan.
	// If the member is in the database but not in the plan, remove it.
	for _, memberInDB := range membersInDB {
		found := false
		for _, memberInPlan := range state.Members {
			if memberInPlan.ValueString() == memberInDB.Name {
				found = true
			}
		}
		if !found {
			login := &qmodel.Login{
				Name: memberInDB.Name,
			}
			loginsToRemove = append(loginsToRemove, login)
		}
	}

	// Add the members to the role
	err = r.connector.AddServerRoleMembers(dbCtx, db, role, loginsToAdd)
	if err != nil {
		resp.Diagnostics.AddError("Error adding logins to role", err.Error())
		return
	}

	// Remove the members from the role
	err = r.connector.RemoveServerRoleMembers(dbCtx, db, role, loginsToRemove)
	if err != nil {
		resp.Diagnostics.AddError("Error removing logins from role", err.Error())
		return
	}

	tflog.Debug(ctx, "ServerRoleResource: populate the state object (model.RoleModel) ")
	state.Name = types.StringValue(role.Name)
	state.PrincipalID = types.Int64Value(role.PrincipalID)
	state.Type = types.StringValue(role.Type)
	state.TypeDescription = types.StringValue(role.TypeDescription)
	state.OwningPrincipal = types.StringValue(role.OwningPrincipal)
	state.IsFixedRole = types.BoolValue(role.IsFixedRole)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// ImportState implements resource.ResourceWithImportState.
func (r *ServerRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	panic("not implemented")
}
