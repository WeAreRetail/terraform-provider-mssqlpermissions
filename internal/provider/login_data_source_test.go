package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLoginDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccLoginDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssqlpermissions_login.test", "id", "1"),
				),
			},
		},
	})
}

const testAccLoginDataSourceConfig = `
data "mssqlpermissions_login" "test" {
	config = {
		server_fqdn   = "mssql-fixture"
		server_port   = 1433
		database_name = "ApplicationDB"
	
		sql_login = {
		  username = "sa"
		  password = "P@ssw0rd"
		}
	}

    name = "sa"
}
`
