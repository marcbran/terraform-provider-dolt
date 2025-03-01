package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccViewResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() +
					testAccDatabaseResourceConfig() +
					testAccTableResourceConfig() +
					testAccViewResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testAccViewResourceConfig() string {
	return `
resource "dolt_view" "test" {
  database = dolt_database.test.name

  name  = "test_view"
  query = <<EOF
SELECT name FROM test_table
EOF
}
`
}
