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

var _ resource.Resource = &DatabaseRoleMembersResource{}
var _ resource.ResourceWithImportState = &DatabaseRoleMembersResource{}
var _ resource.ResourceWithConfigure = &DatabaseRoleMembersResource{}

func NewDatabaseRoleMembersResource() resource.Resource {
	return &DatabaseRoleMembersResource{}
}

type DatabaseRoleMembersResource struct {
	connector *queries.Connector
}

// Metadata is a method that sets the metadata for the DatabaseRoleMembersResource.
// It takes a context.Context, a resource.MetadataRequest, and a pointer to a resource.MetadataResponse as parameters.
// It sets the TypeName field of the MetadataResponse to the concatenation of the ProviderTypeName from the MetadataRequest and "_database_role".
func (r *DatabaseRoleMembersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_role_members"
}

// Schema is a method that sets the schema for the DatabaseRoleMembersResource.
// It defines the attributes and their descriptions for the resource.
func (r *DatabaseRoleMembersResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"members": schema.ListAttribute{
				Description:         "The database role's members.",
				MarkdownDescription: "The database role's members.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure adds the provider-configured client to the resource.
func (r *DatabaseRoleMembersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create is a method of the DatabaseRoleMembersResource struct that creates a new database role.
// It takes a context.Context, a resource.CreateRequest, and a pointer to a resource.CreateResponse as parameters.
// It connects to the database, creates the role, retrieves the created role, and adds members to the role.
// It updates the state object with the created role information.
// If any error occurs during the process, it adds the error to the response diagnostics.
func (r *DatabaseRoleMembersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state model.RoleMembersModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "DatabaseRoleMembersResource", "Create")

	// Use provider connector
	connector := r.connector

	// Connect to database using helper function
	db, err := connectToDatabase(ctx, connector)
	if err != nil {
		handleDatabaseConnectionError(ctx, err, &resp.Diagnostics)
		return
	}

	// Get the role first
	role := &qmodel.Role{
		Name: state.Name.ValueString(),
	}

	role, err = connector.GetDatabaseRole(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving the role", err.Error())
		return
	}

	// Add members to the role
	// Convert members list to slice for processing
	members, convertDiags := convertStringListToSlice(ctx, state.Members)
	if convertDiags != nil {
		resp.Diagnostics.Append(*convertDiags...)
		return
	}

	for _, memberName := range members {
		user := &qmodel.User{
			Name: memberName,
		}

		err = connector.AddDatabaseRoleMember(ctx, db, role, user)
		if err != nil {
			resp.Diagnostics.AddError("Error adding user to role", err.Error())
			return
		}
	}

	// Set state
	state.Name = types.StringValue(role.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	logResourceOperationComplete(ctx, "DatabaseRoleMembersResource", "Create")
}

// Delete deletes database role members.
//
// It connects to the database using the provided connector, retrieves the role information from the state,
// and deletes the role from the database.
//
// If there is an error connecting to the database or deleting the role, it adds an error diagnostic to the response.
func (r *DatabaseRoleMembersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.RoleMembersModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "DatabaseRoleMembersResource", "Delete")

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

	members, err := connector.GetDatabaseRoleMembers(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error getting role members", err.Error())
		return
	}

	// Remove the members from the role
	err = connector.RemoveDatabaseRoleMembers(ctx, db, role, members)
	if err != nil {
		resp.Diagnostics.AddError("Error removing role members", err.Error())
		return
	}

	logResourceOperationComplete(ctx, "DatabaseRoleMembersResource", "Delete")
}

// Read reads the state of the DatabaseRoleMembersResource.
// It retrieves the role information from the database and populates the state object.
// If the role is not found, it creates an empty state object.
// It returns any diagnostics encountered during the process.
func (r *DatabaseRoleMembersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.RoleMembersModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "DatabaseRoleMembersResource", "Read")

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

	// Get the members of the role
	members, err := connector.GetDatabaseRoleMembers(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error getting role members", err.Error())
		return
	}

	// ⚠️ We need to keep the same order as the state, else terraform will detect a change.

	// Convert state members to slice for processing
	stateMembers, convertDiags := convertStringListToSlice(ctx, state.Members)
	if convertDiags != nil {
		resp.Diagnostics.Append(*convertDiags...)
		return
	}

	// List the users in the State and create a list with the users in the database in the same order.
	// After this step, we have ordered the user in the database the same way as the user in the state. But additional users from the database still need to be added.
	var futureStateMembers []string
	for _, stateMemberName := range stateMembers {
		for _, currentMember := range members {
			if currentMember.Name == stateMemberName {
				futureStateMembers = append(futureStateMembers, currentMember.Name)
				break
			}
		}
	}

	// Add in futureStateMembers all the users in the "members" list but not yet in the futureStateMembers list.
	for _, currentMember := range members {
		found := false
		for _, stateMemberName := range stateMembers {
			if currentMember.Name == stateMemberName {
				found = true
				break
			}
		}
		if !found && currentMember.Name != "dbo" { // Ignore "dbo" as it is a special user that cannot be managed
			futureStateMembers = append(futureStateMembers, currentMember.Name)
		}
	}

	// Convert back to types.List
	futureStateList, convertDiags := convertStringSliceToList(ctx, futureStateMembers)
	if convertDiags != nil {
		resp.Diagnostics.Append(*convertDiags...)
		return
	}

	state.Name = types.StringValue(role.Name)
	state.Members = futureStateList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	logResourceOperationComplete(ctx, "DatabaseRoleMembersResource", "Read")
}

// Update updates the database role based on the provided update request.
// Update only the members of the role.
// It compares the members in the plan with the members in the database and adds or removes members accordingly.
// It also populates the state object with the updated role information.

func (r *DatabaseRoleMembersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state model.RoleMembersModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResourceOperation(ctx, "DatabaseRoleMembersResource", "Update")

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
	if err != nil && err.Error() != "database role not found" {
		resp.Diagnostics.AddError("Error getting role", err.Error())
		return
	}

	membersInDB, err := connector.GetDatabaseRoleMembers(ctx, db, role)
	if err != nil {
		resp.Diagnostics.AddError("Error getting role members", err.Error())
		return
	}

	var usersToAdd = make([]*qmodel.User, 0)
	var usersToRemove = make([]*qmodel.User, 0)

	// Convert state members to slice for processing
	stateMembers, convertDiags := convertStringListToSlice(ctx, state.Members)
	if convertDiags != nil {
		resp.Diagnostics.Append(*convertDiags...)
		return
	}

	// Compare the members in the plan with the members in the database.
	// If the member is in the plan but not in the database, add it.
	for _, memberName := range stateMembers {
		found := false
		for _, memberInDB := range membersInDB {
			if memberName == memberInDB.Name {
				found = true
			}
		}
		if !found {
			user := &qmodel.User{
				Name: memberName,
			}
			usersToAdd = append(usersToAdd, user)
		}
	}

	// Compare the members in the database with the members in the plan.
	// If the member is in the database but not in the plan, remove it.
	for _, memberInDB := range membersInDB {
		found := false
		for _, memberName := range stateMembers {
			if memberName == memberInDB.Name {
				found = true
			}
		}
		if !found {
			// Ignore "dbo" as it is a special user that cannot be managed
			if memberInDB.Name == "dbo" {
				continue
			}

			user := &qmodel.User{
				Name: memberInDB.Name,
			}
			usersToRemove = append(usersToRemove, user)
		}
	}

	// Add the members to the role
	err = connector.AddDatabaseRoleMembers(ctx, db, role, usersToAdd)
	if err != nil {
		resp.Diagnostics.AddError("Error adding users to role", err.Error())
		return
	}

	// Remove the members from the role
	err = connector.RemoveDatabaseRoleMembers(ctx, db, role, usersToRemove)
	if err != nil {
		resp.Diagnostics.AddError("Error removing users from role", err.Error())
		return
	}

	state.Name = types.StringValue(role.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	logResourceOperationComplete(ctx, "DatabaseRoleMembersResource", "Update")
}

// ImportState implements resource.ResourceWithImportState.
func (r *DatabaseRoleMembersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	panic("not implemented")
}
