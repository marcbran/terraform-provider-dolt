package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatabaseResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() +
					testAccDatabaseResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testAccDatabaseResourceConfig() string {
	return `
resource "dolt_database" "test" {
  name = "test"
}
`
}
