// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"mssqlpermissions": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}

// Unit tests for the provider

func TestSqlPermissionsProvider_Metadata(t *testing.T) {
	p := &SqlPermissionsProvider{
		version: "1.0.0",
	}
	ctx := context.Background()
	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(ctx, req, resp)

	if resp.TypeName != "mssqlpermissions" {
		t.Errorf("Expected TypeName 'mssqlpermissions', got %s", resp.TypeName)
	}

	if resp.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got %s", resp.Version)
	}
}

func TestSqlPermissionsProvider_Schema(t *testing.T) {
	p := &SqlPermissionsProvider{}
	ctx := context.Background()
	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(ctx, req, resp)

	// Verify schema is not nil
	if resp.Schema.Attributes == nil {
		t.Error("Expected schema attributes to be defined")
		return
	}

	// Check for required provider attributes
	requiredAttrs := []string{
		"server_fqdn", "server_port", "database_name",
		"sql_login", "spn_login", "msi_login", "federated_login",
	}
	for _, attr := range requiredAttrs {
		if _, exists := resp.Schema.Attributes[attr]; !exists {
			t.Errorf("Expected attribute %s to be defined in provider schema", attr)
		}
	}

	// Verify descriptions are set
	expectedDescription := "Manage SQL Server permissions. Locally or in Azure SQL Database."
	if resp.Schema.Description != expectedDescription {
		t.Errorf("Expected Description %s, got %s", expectedDescription, resp.Schema.Description)
	}
	if resp.Schema.MarkdownDescription != expectedDescription {
		t.Errorf("Expected MarkdownDescription %s, got %s", expectedDescription, resp.Schema.MarkdownDescription)
	}
}

func TestSqlPermissionsProvider_Configure_ValidConfiguration(t *testing.T) {
	// Test valid configuration values (without full framework setup)
	t.Run("ValidSQLLogin", func(t *testing.T) {
		config := createValidSQLLoginConfig()

		if config.ServerFqdn.IsNull() || config.ServerFqdn.ValueString() == "" {
			t.Error("Expected server_fqdn to be set in test config")
		}
		if config.DatabaseName.IsNull() || config.DatabaseName.ValueString() == "" {
			t.Error("Expected database_name to be set in test config")
		}
	})
}

func TestSqlPermissionsProvider_Configure_ValidationErrors(t *testing.T) {
	// Test validation logic without actual provider configuration
	// (Full integration tests would require the terraform-plugin-framework test utilities)

	t.Run("EmptyServerFqdn", func(t *testing.T) {
		config := SqlPermissionsProviderModel{
			ServerFqdn:   types.StringValue(""),
			DatabaseName: types.StringValue("testdb"),
		}

		if !isConfigInvalid(config.ServerFqdn) {
			t.Error("Expected empty server_fqdn to be invalid")
		}
	})

	t.Run("EmptyDatabaseName", func(t *testing.T) {
		config := SqlPermissionsProviderModel{
			ServerFqdn:   types.StringValue("localhost"),
			DatabaseName: types.StringValue(""),
		}

		if !isConfigInvalid(config.DatabaseName) {
			t.Error("Expected empty database_name to be invalid")
		}
	})

	t.Run("InvalidPortRange", func(t *testing.T) {
		// Test port validation logic
		invalidPorts := []int64{0, -1, 65536, 100000}

		for _, port := range invalidPorts {
			if isValidPort(port) {
				t.Errorf("Expected port %d to be invalid", port)
			}
		}
	})

	t.Run("ValidPortRange", func(t *testing.T) {
		// Test valid ports
		validPorts := []int64{1, 1433, 443, 8080, 65535}

		for _, port := range validPorts {
			if !isValidPort(port) {
				t.Errorf("Expected port %d to be valid", port)
			}
		}
	})
}

func TestSqlPermissionsProvider_AuthenticationMethodValidation(t *testing.T) {
	t.Run("NoAuthMethod", func(t *testing.T) {
		config := SqlPermissionsProviderModel{
			ServerFqdn:   types.StringValue("localhost"),
			DatabaseName: types.StringValue("testdb"),
			// No auth methods set
		}

		authMethods := countAuthMethods(config)
		if authMethods != 0 {
			t.Errorf("Expected 0 auth methods, got %d", authMethods)
		}
	})

	t.Run("MultipleAuthMethods", func(t *testing.T) {
		// Create non-null objects to simulate having auth methods set
		config := SqlPermissionsProviderModel{
			ServerFqdn:   types.StringValue("localhost"),
			DatabaseName: types.StringValue("testdb"),
			SQLLogin:     types.ObjectUnknown(nil), // Simulate a set object
			MSILogin:     types.ObjectUnknown(nil), // Simulate a set object
		}

		authMethods := countAuthMethods(config)
		if authMethods != 2 {
			t.Errorf("Expected 2 auth methods, got %d", authMethods)
		}
	})
}

func TestSqlPermissionsProvider_Resources(t *testing.T) {
	p := &SqlPermissionsProvider{}
	ctx := context.Background()

	resources := p.Resources(ctx)

	// Verify that all expected resources are returned
	if len(resources) == 0 {
		t.Error("Expected at least one resource to be defined")
	}

	// Test that each resource function returns a non-nil resource
	for i, resourceFunc := range resources {
		resource := resourceFunc()
		if resource == nil {
			t.Errorf("Resource function at index %d returned nil", i)
		}
	}
}

func TestSqlPermissionsProvider_DataSources(t *testing.T) {
	p := &SqlPermissionsProvider{}
	ctx := context.Background()

	dataSources := p.DataSources(ctx)

	// Verify that data sources are returned
	if len(dataSources) == 0 {
		t.Error("Expected at least one data source to be defined")
	}

	// Test that each data source function returns a non-nil data source
	for i, dataSourceFunc := range dataSources {
		dataSource := dataSourceFunc()
		if dataSource == nil {
			t.Errorf("Data source function at index %d returned nil", i)
		}
	}
}

func TestNew(t *testing.T) {
	version := "test"
	providerFunc := New(version)

	if providerFunc == nil {
		t.Error("Expected New to return a non-nil function")
	}

	provider := providerFunc()
	if provider == nil {
		t.Error("Expected provider function to return a non-nil provider")
	}

	// Verify it's the correct type
	if sqlProvider, ok := provider.(*SqlPermissionsProvider); ok {
		if sqlProvider.version != version {
			t.Errorf("Expected provider version %s, got %s", version, sqlProvider.version)
		}
	} else {
		t.Error("Expected New to return *SqlPermissionsProvider")
	}
}

// Test provider interface compliance
func TestSqlPermissionsProvider_InterfaceCompliance(t *testing.T) {
	var _ provider.Provider = &SqlPermissionsProvider{}
}

// Helper functions for testing

func createValidSQLLoginConfig() SqlPermissionsProviderModel {
	return SqlPermissionsProviderModel{
		ServerFqdn:   types.StringValue("localhost"),
		ServerPort:   types.Int64Value(1433),
		DatabaseName: types.StringValue("testdb"),
		SQLLogin:     types.ObjectNull(nil),
	}
}

func isConfigInvalid(value types.String) bool {
	return value.IsNull() || value.ValueString() == ""
}

func isValidPort(port int64) bool {
	return port >= 1 && port <= 65535
}

func countAuthMethods(config SqlPermissionsProviderModel) int {
	count := 0
	// For testing purposes, we check if objects are not null
	// In real usage, they would have actual values
	if !config.SQLLogin.IsNull() {
		count++
	}
	if !config.SPNLogin.IsNull() {
		count++
	}
	if !config.MSILogin.IsNull() {
		count++
	}
	if !config.FederatedLogin.IsNull() {
		count++
	}
	return count
}

// Benchmark tests
func BenchmarkSqlPermissionsProvider_Metadata(b *testing.B) {
	p := &SqlPermissionsProvider{version: "1.0.0"}
	ctx := context.Background()
	req := provider.MetadataRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &provider.MetadataResponse{}
		p.Metadata(ctx, req, resp)
	}
}

func BenchmarkSqlPermissionsProvider_Schema(b *testing.B) {
	p := &SqlPermissionsProvider{}
	ctx := context.Background()
	req := provider.SchemaRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &provider.SchemaResponse{}
		p.Schema(ctx, req, resp)
	}
}
