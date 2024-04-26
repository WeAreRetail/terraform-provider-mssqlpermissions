// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserResourceConfig("one", "P@ssw0rd"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_user.test", "name", "one"),
					resource.TestCheckResourceAttr("mssqlpermissions_user.test", "password", "P@ssw0rd"),
				),
			},
			// Update and Read testing
			{
				Config: testAccUserResourceConfig("one", "P@ssw0rd!"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssqlpermissions_user.test", "password", "P@ssw0rd!"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccUserResourceConfig(name string, password string) string {
	return fmt.Sprintf(`
resource "mssqlpermissions_user" "test" {
	config = {
		server_fqdn   = %q
		server_port   = %q
		database_name = "ApplicationDB"

		sql_login = {
		  username = "sa"
		  password = "P@ssw0rd"
		}
	}

	name     = %q
	password = %q
}
`, os.Getenv("LOCAL_SQL_HOST"), os.Getenv("LOCAL_SQL_PORT"), name, password)
}
