// Centralized configuration for the database connection.
// Used by every resource and data source.

package provider

import (
	"context"
	"queries"
	"terraform-provider-mssqlpermissions/internal/provider/model"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// getConfigSchema returns the schema for the database connection configuration.
// It defines the attributes for configuring the SQL Server connection, including server FQDN, port, database name,
// SQL Server login credentials, Service Principal Name (SPN) credentials, Managed Identity credentials, and Federated Identity.
// The attributes are marked as optional or required based on their usage.
func getConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description:         "The database connection configuration",
		MarkdownDescription: "The database connection configuration",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"server_fqdn": schema.StringAttribute{
				Description:         "The SQL Server FQDN.",
				MarkdownDescription: "The SQL Server FQDN.",
				Required:            true,
			},
			"server_port": schema.Int64Attribute{
				Description:         "The SQL Server port.",
				MarkdownDescription: "The SQL Server port.",
				Optional:            true,
			},
			"database_name": schema.StringAttribute{
				Description:         "The SQL Server database name.",
				MarkdownDescription: "The SQL Server database name.",
				Required:            true,
			},
			"sql_login": schema.SingleNestedAttribute{
				Description:         "The SQL Server login configuration. Use to connect to the Database using SQL Authentication.",
				MarkdownDescription: "The SQL Server login configuration. Use to connect to the Database using SQL Authentication.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Description:         "The SQL Server login username.",
						MarkdownDescription: "The SQL Server login username.",
						Required:            true,
					},
					"password": schema.StringAttribute{
						Description:         "The SQL Server login password.",
						MarkdownDescription: "The SQL Server login password.",
						Required:            true,
					},
				},
			},
			"spn_login": schema.SingleNestedAttribute{
				Description:         "Connect using a Service Principal Name (SPN).",
				MarkdownDescription: "Connect using a Service Principal Name (SPN).",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{
						Description:         "The Azure AD application client ID.",
						MarkdownDescription: "The Azure AD application client ID.",
						Required:            true,
					},
					"client_secret": schema.StringAttribute{
						Description:         "The Azure AD application client secret.",
						MarkdownDescription: "The Azure AD application client secret.",
						Required:            true,
					},
					"tenant_id": schema.StringAttribute{
						Description:         "The Azure AD tenant ID.",
						MarkdownDescription: "The Azure AD tenant ID.",
						Required:            true,
					},
				},
			},
			"msi_login": schema.SingleNestedAttribute{
				Description:         "Connect using a Managed Identity.",
				MarkdownDescription: "Connect using a Managed Identity.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"user_identity": schema.BoolAttribute{
						Description:         "Use the user identity.",
						MarkdownDescription: "Use the user identity.",
						Required:            true,
					},
					"user_id": schema.StringAttribute{
						Description:         "The user identity. Required if user_identity is true.",
						MarkdownDescription: "The user identity. Required if user_identity is true.",
						Optional:            true,
					},
					"resource_id": schema.StringAttribute{
						Description:         "The resource identity. Required if user_identity is false.",
						MarkdownDescription: "The resource identity. Required if user_identity is false.",
						Optional:            true,
					},
				},
			},
			"federated_login": schema.SingleNestedAttribute{
				Description:         "Connect using a Federated Identity",
				MarkdownDescription: "Connect using a Federated Identity",
				Optional:            true,
			},
		},
	}
}

// getConnector returns the connector for the database connection configuration.
func getConnector(config *model.ConfigModel) (*queries.Connector, diag.Diagnostics) {

	ctx := context.Background()
	var sqlLogin model.SQLLoginModel
	var spnLogin model.SPNLoginModel
	var msiLogin model.MSILoginModel

	connector := &queries.Connector{
		Host:     config.ServerFqdn.ValueString(),
		Port:     int(config.ServerPort.ValueInt64()),
		Database: config.DatabaseName.ValueString(),
	}

	if !config.SQLLogin.IsNull() && !config.SQLLogin.IsUnknown() {
		diags := config.SQLLogin.As(ctx, &sqlLogin, basetypes.ObjectAsOptions{})

		if diags.HasError() {
			return nil, diags
		}

		connector.LocalUserLogin = &queries.LocalUserLogin{
			Username: sqlLogin.Username.ValueString(),
			Password: sqlLogin.Password.ValueString(),
		}
	}

	if !config.SPNLogin.IsNull() && !config.SPNLogin.IsUnknown() {

		diags := config.SPNLogin.As(ctx, &spnLogin, basetypes.ObjectAsOptions{})

		if diags.HasError() {
			return nil, diags
		}

		connector.AzureApplicationLogin = &queries.AzureApplicationLogin{
			ClientId:     spnLogin.ClientID.ValueString(),
			ClientSecret: spnLogin.ClientSecret.ValueString(),
			TenantId:     spnLogin.TenantID.ValueString(),
		}
	}

	if !config.MSILogin.IsNull() && !config.MSILogin.IsUnknown() {

		diags := config.MSILogin.As(ctx, &msiLogin, basetypes.ObjectAsOptions{})

		if diags.HasError() {
			return nil, diags
		}

		connector.ManagedIdentityLogin = &queries.ManagedIdentityLogin{
			UserIdentity: msiLogin.UserIdentity.ValueBool(),
			UserId:       msiLogin.UserId.ValueString(),
			ResourceId:   msiLogin.ResourceId.ValueString(),
		}
	}

	return connector, nil
}
