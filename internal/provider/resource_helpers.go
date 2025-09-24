package provider

import (
	"context"
	"database/sql"
	"terraform-provider-mssqlpermissions/internal/provider/model"
	"terraform-provider-mssqlpermissions/internal/queries"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// getResourceConnector gets the connector for a resource, using provider-level config if available,
// otherwise falling back to resource-level config for backward compatibility.
func getResourceConnector(ctx context.Context, providerConnector *queries.Connector, resourceConfig *model.ConfigModel) (*queries.Connector, diag.Diagnostics) {
	tflog.Debug(ctx, "Getting resource connector")

	// Use provider-level connector if available
	if providerConnector != nil {
		tflog.Debug(ctx, "Using provider-level connector")
		return providerConnector, nil
	}

	// Fall back to resource-level configuration for backward compatibility
	if resourceConfig != nil {
		tflog.Debug(ctx, "Falling back to resource-level configuration")
		return getConnector(resourceConfig)
	}

	var diags diag.Diagnostics
	diags.AddError(
		"No Database Configuration Found",
		"Neither provider-level nor resource-level database configuration was found. "+
			"Please configure the provider with database connection details.",
	)
	return nil, diags
}

// connectToDatabase establishes a database connection using the provided connector and context.
func connectToDatabase(ctx context.Context, connector *queries.Connector) (*sql.DB, error) {
	tflog.Debug(ctx, "Connecting to database")
	return connector.Connect()
}

// handleDatabaseConnectionError is a standardized error handler for database connection failures.
func handleDatabaseConnectionError(ctx context.Context, err error, diags *diag.Diagnostics) {
	if err == nil {
		return
	}

	tflog.Error(ctx, "Database connection failed", map[string]interface{}{
		"error": err.Error(),
	})
	diags.AddError("Database Connection Failed", err.Error())
}

// logResourceOperation logs the start of a resource operation for debugging.
func logResourceOperation(ctx context.Context, resourceType, operation string) {
	tflog.Debug(ctx, "Resource operation started", map[string]interface{}{
		"resource_type": resourceType,
		"operation":     operation,
	})
}

// logResourceOperationComplete logs the completion of a resource operation for debugging.
func logResourceOperationComplete(ctx context.Context, resourceType, operation string) {
	tflog.Debug(ctx, "Resource operation completed", map[string]interface{}{
		"resource_type": resourceType,
		"operation":     operation,
	})
}
