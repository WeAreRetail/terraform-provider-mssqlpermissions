package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccUserDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssqlpermissions_user.test", "principal_id", "1"),
				),
			},
		},
	})
}

func testAccUserDataSourceConfig() string {
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

data "mssqlpermissions_user" "test" {
	name = "dbo"
}
`, os.Getenv("LOCAL_SQL_HOST"), os.Getenv("LOCAL_SQL_PORT"))
}
