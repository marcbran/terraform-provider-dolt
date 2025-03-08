package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRowSetResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() +
					testAccDatabaseResourceConfig() +
					testAccTableResourceConfig() +
					testAccRowSetResourceConfigOne(),
				Check: resource.ComposeAggregateTestCheckFunc(),
			},
			{
				Config: testAccProviderConfig() +
					testAccDatabaseResourceConfig() +
					testAccTableResourceConfig() +
					testAccRowSetResourceConfigTwo(),
				Check: resource.ComposeAggregateTestCheckFunc(),
			},
			{
				Config: testAccProviderConfig() +
					testAccDatabaseResourceConfig() +
					testAccTableResourceConfig() +
					testAccRowSetResourceConfigOne(),
				Check: resource.ComposeAggregateTestCheckFunc(),
			},
			{
				Config: testAccProviderConfig() +
					testAccDatabaseResourceConfig() +
					testAccTableResourceConfig() +
					testAccRowSetResourceConfigZero(),
				Check: resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testAccRowSetResourceConfigZero() string {
	return `
resource "dolt_rowset" "test" {
  database = dolt_database.test.name
  table    = dolt_table.test.name

  columns       = ["id", "name"]
  unique_column = "id"
  values        = {}
}
`
}

func testAccRowSetResourceConfigOne() string {
	return `
resource "dolt_rowset" "test" {
  database = dolt_database.test.name
  table    = dolt_table.test.name

  columns       = ["id", "name"]
  unique_column = "id"
  values  = {
    1 = ["1", "Alice"],
  }
}
`
}

func testAccRowSetResourceConfigTwo() string {
	return `
resource "dolt_rowset" "test" {
  database = dolt_database.test.name
  table    = dolt_table.test.name

  columns       = ["id", "name"]
  unique_column = "id"
  values  = {
    1 = ["1", "Alice"],
    2 = ["2", "Bob"],
  }
}
`
}
