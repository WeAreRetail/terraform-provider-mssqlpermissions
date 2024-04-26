// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServerRoleResourceLocal(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccServerRoleResourceConfigLocalSQL("one", "\"loginFixtureOne\",\"loginFixtureTwo\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_server_role.test", "name", "one"),
					resource.TestCheckResourceAttr("mssqlpermissions_server_role.test", "owning_principal", "1"),
				),
			},
			// Update and Read testing
			{
				Config: testAccServerRoleResourceConfigLocalSQL("two", "\"loginFixtureOne\",\"loginFixtureTwo\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_server_role.test", "name", "two"),
				),
			},
			// Update and Read testing
			{
				Config: testAccServerRoleResourceConfigLocalSQL("two", "\"loginFixtureTwo\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_server_role.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccServerRoleResourceConfigLocalSQL(name string, member string) string {
	return fmt.Sprintf(`
resource "mssqlpermissions_server_role" "test" {
	config = {
		server_fqdn   = %q
		server_port   = %q
		database_name = "master"

		sql_login = {
		  username = "sa"
		  password = "P@ssw0rd"
		}
	  }

	name     = %q
	members  = [%s]
}
`, os.Getenv("LOCAL_SQL_HOST"), os.Getenv("LOCAL_SQL_PORT"), name, member)
}
