// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
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
				Config: testAccDatabaseRoleResourceConfigLocalSQL("one", "\"userFixtureOne\",\"userFixtureTwo\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "name", "one"),
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "owning_principal", "1"),
				),
			},
			// Update and Read testing
			{
				Config: testAccDatabaseRoleResourceConfigLocalSQL("two", "\"userFixtureOne\",\"userFixtureTwo\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "name", "two"),
				),
			},
			// Update and Read testing
			{
				Config: testAccDatabaseRoleResourceConfigLocalSQL("two", "\"userFixtureTwo\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDatabaseRoleResourceConfigLocalSQL(name string, member string) string {
	return fmt.Sprintf(`
resource "mssqlpermissions_database_role" "test" {
	config = {
		server_fqdn   = "mssql-fixture"
		server_port   = 1433
		database_name = "ApplicationDB"
	
		sql_login = {
		  username = "sa"
		  password = "P@ssw0rd"
		}
	  }

	name     = %q
	members  = [%s]
}
`, name, member)
}

func TestAccDatabaseRoleResourceAzure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDatabaseRoleResourceConfigAzureSQL("one", "\"userFixtureOne\",\"fixtureTwo\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "name", "one"),
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "owning_principal", "1"),
				),
			},
			// Update and Read testing
			{
				Config: testAccDatabaseRoleResourceConfigAzureSQL("two", "\"userFixtureOne\",\"fixtureTwo\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "name", "two"),
				),
			},
			// Update and Read testing
			{
				Config: testAccDatabaseRoleResourceConfigAzureSQL("two", "\"userFixtureTwo\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_database_role.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDatabaseRoleResourceConfigAzureSQL(name string, member string) string {
	return fmt.Sprintf(`
resource "mssqlpermissions_database_role" "test" {
	config = {
		server_fqdn   = "d10tfp76c9sqlvrle.database.windows.net"
		server_port   = 1433
		database_name = "ApplicationDB"
	  }

	name     = %q
	members  = [%s]
}
`, name, member)
}
