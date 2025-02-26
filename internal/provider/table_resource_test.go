package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTableResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(".") +
					testAccDatabaseResourceConfig() +
					testAccTableResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testAccTableResourceConfig() string {
	return fmt.Sprintf(`
resource "dolt_table" "test" {
  database = dolt_database.test.name

  name  = "test_table"
  query = <<EOF
CREATE TABLE test_table (
	id INT PRIMARY KEY,
	name VARCHAR(100)
);
EOF
}
`)
}
