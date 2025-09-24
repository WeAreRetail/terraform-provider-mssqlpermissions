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

			// TODO: Add external deletion test scenario:
			// 1. Create resource with Terraform
			// 2. Use a custom test step that manually deletes the resource from the database
			// 3. Plan should show resource will be recreated (not deleted)
			// 4. Apply should succeed without "cannot delete... does not exist" errors
			// This would catch the bug we just fixed where external deletion wasn't handled properly.
		},
	})
}

func testAccUserResourceConfig(name string, password string) string {
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

resource "mssqlpermissions_user" "test" {
	name     = %q
	password = %q
}
`, os.Getenv("LOCAL_SQL_HOST"), os.Getenv("LOCAL_SQL_PORT"), name, password)
}
