// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatabaseRoleResourceLocal(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabaseRoleResourceConfigLocalSQL("one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "name", "one"),
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "owning_principal", "1"),
				),
			},
			// Update and Read testing
			{
				Config: testAccDatabaseRoleResourceConfigLocalSQL("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "name", "two"),
				),
			},
			// Update and Read testing
			{
				Config: testAccDatabaseRoleResourceConfigLocalSQL("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDatabaseRoleResourceConfigLocalSQL(name string) string {
	return fmt.Sprintf(`
provider "mssqlpermissions" {
	server_fqdn   = %q
	server_port   = %q
	database_name = "ApplicationDB"

	sql_login = {
		username = "sa"
		password = "P@ssw0rd"
	}
}

resource "mssqlpermissions_database_role" "test" {
	name     = %q
}
`, os.Getenv("LOCAL_SQL_HOST"), os.Getenv("LOCAL_SQL_PORT"), name)
}

func TestAccDatabaseRoleResourceAzure(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip()
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabaseRoleResourceConfigAzureSQL("one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "name", "one"),
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "owning_principal", "1"),
				),
			},
			// Update and Read testing
			{
				Config: testAccDatabaseRoleResourceConfigAzureSQL("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "name", "two"),
				),
			},
			// Update and Read testing
			{
				Config: testAccDatabaseRoleResourceConfigAzureSQL("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDatabaseRoleResourceConfigAzureSQL(name string) string {
	return fmt.Sprintf(`
provider "mssqlpermissions" {
	server_fqdn   = %q
	server_port   = %q
	database_name = "ApplicationDB"

	sql_login = {
		username = %q
		password = %q
	}
}

resource "mssqlpermissions_database_role" "test" {
	name     = %q
}
`, os.Getenv("AZURE_MSSQL_SERVER"), os.Getenv("MSSQL_PORT"), os.Getenv("AZURE_MSSQL_USERNAME"), os.Getenv("AZURE_MSSQL_PASSWORD"), name)
}
