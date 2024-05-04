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

var _ resource.Resource = &DatabaseRoleResource{}
var _ resource.ResourceWithImportState = &DatabaseRoleResource{}

func NewDatabaseRoleResource() resource.Resource {
	return &DatabaseRoleResource{}
}

type DatabaseRoleResource struct {
	connector *queries.Connector
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
			"config": getConfigSchema(), // config is the configuration block shared by all resources and data sources.

			"name": schema.StringAttribute{
				Description:         "The database role's name.",
				MarkdownDescription: "The database role's name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"members": schema.ListAttribute{
				Description:         "The database role's members.",
				MarkdownDescription: "The database role's members.",
				Optional:            true,
				ElementType:         types.StringType,
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
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "DatabaseRoleResource: getConnector")
	r.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "DatabaseRoleResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	role := &qmodel.Role{
		Name: state.Name.ValueString(),
	}

	tflog.Debug(ctx, "DatabaseRoleResource: create the role")
	err = r.connector.CreateDatabaseRole(dbCtx, db, role)

	if err != nil {
		resp.Diagnostics.AddError("Error creating role", err.Error())
		return
	}

	tflog.Debug(ctx, "DatabaseRoleResource: get the created role")
	role, err = r.connector.GetDatabaseRole(dbCtx, db, role)

	if err != nil {
		resp.Diagnostics.AddError("Error retrieving the created role", err.Error())
		return
	}

	tflog.Debug(ctx, "userDataSource: populate the state object (model.UserModel) ")
	state.Name = types.StringValue(role.Name)
	state.PrincipalID = types.Int64Value(role.PrincipalID)
	state.Type = types.StringValue(role.Type)
	state.TypeDescription = types.StringValue(role.TypeDescription)
	state.OwningPrincipal = types.StringValue(role.OwningPrincipal)
	state.IsFixedRole = types.BoolValue(role.IsFixedRole)

	// Add the members to the role
	for _, member := range state.Members {
		user := &qmodel.User{
			Name: member.ValueString(),
		}

		tflog.Debug(ctx, "DatabaseRoleResource: add the member to the role")
		err = r.connector.AddDatabaseRoleMember(dbCtx, db, role, user)
		if err != nil {
			resp.Diagnostics.AddError("Error adding user to role", err.Error())
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes a database role.
//
// It connects to the database using the provided connector, retrieves the role information from the state,
// and deletes the role from the database.
//
// If there is an error connecting to the database or deleting the role, it adds an error diagnostic to the response.
func (r *DatabaseRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.RoleModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "DatabaseRoleResource: getConnector")
	r.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "DatabaseRoleResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	role := &qmodel.Role{
		Name: state.Name.ValueString(),
	}

	tflog.Debug(ctx, "DatabaseRoleResource: create the role")
	err = r.connector.DeleteDatabaseRole(dbCtx, db, role)

	if err != nil {
		resp.Diagnostics.AddError("Error deleting role", err.Error())
		return
	}
}

// Read reads the state of the DatabaseRoleResource.
// It retrieves the role information from the database and populates the state object.
// If the role is not found, it creates an empty state object.
// It returns any diagnostics encountered during the process.
func (r *DatabaseRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.RoleModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "DatabaseRoleResource: getConnector")
	r.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "DatabaseRoleResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	role := &qmodel.Role{
		Name: state.Name.ValueString(),
	}

	tflog.Debug(ctx, "DatabaseRoleResource: get the role")
	role, err = r.connector.GetDatabaseRole(dbCtx, db, role)

	if err != nil && err.Error() != "database role not found" {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	tflog.Debug(ctx, "userDataSource: populate the state object (model.UserModel) ")

	if role == nil {
		state = model.RoleModel{
			Config: state.Config,
		}
	} else {
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

// Update updates the database role based on the provided update request.
// Update only the members of the role.
// It compares the members in the plan with the members in the database and adds or removes members accordingly.
// It also populates the state object with the updated role information.

func (r *DatabaseRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state model.RoleModel
	var err error
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "DatabaseRoleResource: getConnector")
	r.connector, diags = getConnector(state.Config)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set up the context and connect to the database.
	dbCtx := context.Background()
	tflog.Debug(ctx, "DatabaseRoleResource: connect to the database")
	db, err := r.connector.Connect()

	if err != nil {
		resp.Diagnostics.AddError("Error connecting to the database", err.Error())
		return
	}

	role := &qmodel.Role{
		Name: state.Name.ValueString(),
	}

	tflog.Debug(ctx, "DatabaseRoleResource: get the role")
	role, err = r.connector.GetDatabaseRole(dbCtx, db, role)
	if err != nil && err.Error() != "database role not found" {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	tflog.Debug(ctx, "DatabaseRoleResource: get the role members")
	membersInDB, err := r.connector.GetDatabaseRoleMembers(dbCtx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error getting role members", err.Error())
		return
	}

	var usersToAdd []*qmodel.User = make([]*qmodel.User, 0)
	var usersToRemove []*qmodel.User = make([]*qmodel.User, 0)

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
			user := &qmodel.User{
				Name: memberInPlan.ValueString(),
			}
			usersToAdd = append(usersToAdd, user)
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
			user := &qmodel.User{
				Name: memberInDB.Name,
			}
			usersToRemove = append(usersToRemove, user)
		}
	}

	// Add the members to the role
	err = r.connector.AddDatabaseRoleMembers(dbCtx, db, role, usersToAdd)
	if err != nil {
		resp.Diagnostics.AddError("Error adding users to role", err.Error())
		return
	}

	// Remove the members from the role
	err = r.connector.RemoveDatabaseRoleMembers(dbCtx, db, role, usersToRemove)
	if err != nil {
		resp.Diagnostics.AddError("Error removing users from role", err.Error())
		return
	}

	tflog.Debug(ctx, "userDataSource: populate the state object (model.UserModel) ")
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
func (r *DatabaseRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	panic("not implemented")
}
