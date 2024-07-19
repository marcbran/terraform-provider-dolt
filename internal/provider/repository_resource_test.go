package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRepositoryResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryResourceConfig("./test"),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testAccRepositoryResourceConfig(path string) string {
	return fmt.Sprintf(`
resource "dolt_repository" "test" {
  path  = %[1]q
  email = "test@example.com"
  name  = "Test Example"
}
`, path)
}
