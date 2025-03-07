package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccTableDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() +
					testAccTableDataSourceConfig(),
				ExpectError: regexp.MustCompile("Cannot find table with name test"),
			},
		},
	})
}

func TestAccExistingTableDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() +
					testAccDatabaseResourceConfig() +
					testAccTableResourceConfig() +
					testAccExistingTableDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.dolt_table.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test_table"),
					),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dolt_table.test", "columns.#", "1"),
					resource.TestCheckResourceAttr("data.dolt_table.test", "columns.0.name", "name"),
					resource.TestCheckResourceAttr("data.dolt_table.test", "columns.0.type", "varchar(100)"),
					resource.TestCheckResourceAttr("data.dolt_table.test", "columns.0.key", ""),
				),
			},
		},
	})
}

func testAccTableDataSourceConfig() string {
	return `
data "dolt_table" "test" {
  database = "test"
  name     = "test"
}
`
}

func testAccExistingTableDataSourceConfig() string {
	return `
data "dolt_table" "test" {
  database = dolt_table.test.database
  name     = dolt_table.test.name
}
`
}
