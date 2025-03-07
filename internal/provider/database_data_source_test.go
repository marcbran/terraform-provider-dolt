package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDatabaseDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() +
					testAccDatabaseDataSourceConfig(),
				ExpectError: regexp.MustCompile("Cannot find database with name test"),
			},
		},
	})
}

func TestAccExistingDatabaseDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() +
					testAccDatabaseResourceConfig() +
					testAccExistingDatabaseDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.dolt_database.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test"),
					),
				},
			},
		},
	})
}

func testAccDatabaseDataSourceConfig() string {
	return `
data "dolt_database" "test" {
  name = "test"
}
`
}

func testAccExistingDatabaseDataSourceConfig() string {
	return `
data "dolt_database" "test" {
  name = dolt_database.test.name
}
`
}
