package provider

import (
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
				Config: testAccUserDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssqlpermissions_user.test", "principal_id", "1"),
				),
			},
		},
	})
}

const testAccUserDataSourceConfig = `
data "mssqlpermissions_user" "test" {
	config = {
		server_fqdn   = "mssql-fixture"
		server_port   = 1433
		database_name = "ApplicationDB"
	
		sql_login = {
		  username = "sa"
		  password = "P@ssw0rd"
		}
	}

    name = "dbo"
}
`
