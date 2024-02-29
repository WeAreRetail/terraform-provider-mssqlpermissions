package provider

import (
	"fmt"
	"os"
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
				Config: testAccLoginDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssqlpermissions_login.test", "id", "1"),
				),
			},
		},
	})
}

func testAccLoginDataSourceConfig() string {
	return fmt.Sprintf(`
data "mssqlpermissions_login" "test" {
	config = {
		server_fqdn   = %q
		server_port   = %q
		database_name = "ApplicationDB"
	
		sql_login = {
			username = "sa"
			password = "P@ssw0rd"
		}
	}

	name = "sa"
}
`, os.Getenv("LOCAL_SQL_HOST"), os.Getenv("LOCAL_SQL_PORT"))
}
