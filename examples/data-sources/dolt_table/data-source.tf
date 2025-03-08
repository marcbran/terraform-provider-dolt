terraform {
  required_providers {
    dolt = {
      source = "marcbran/dolt"
    }
  }
}

provider "dolt" {
  path  = "."
  name  = "John Doe"
  email = "john.doe@example.com"
}

data "dolt_database" "main" {
  name = "main"
}

data "dolt_table" "articles" {
  database = data.dolt_database.main.name
  name     = "articles"
}
