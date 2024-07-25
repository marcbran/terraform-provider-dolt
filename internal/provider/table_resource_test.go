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
				Config: testAccTableResourceConfig("./test"),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testAccTableResourceConfig(path string) string {
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
`, path)
}
