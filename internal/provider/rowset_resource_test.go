package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRowSetResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRowSetResourceConfig("./test"),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testAccRowSetResourceConfig(path string) string {
	return fmt.Sprintf(`
resource "dolt_repository" "test" {
  path  = %[1]q
  email = "test@example.com"
  name  = "Test Example"
}

resource "dolt_table" "test" {
  repository_path = dolt_repository.test.path

  name  = "test_table"
  query = <<EOF
CREATE TABLE test_table (
	id INT PRIMARY KEY,
	name VARCHAR(100)
);
EOF
}

resource "dolt_rowset" "test" {
  repository_path = dolt_repository.test.path
  table_name      = dolt_table.test.name

  columns       = ["id", "name"]
  unique_column = "id"
  values  = {
    1 = ["1", "Alice"],
    2 = ["2", "Bob"],
  }
}
`, path)
}
