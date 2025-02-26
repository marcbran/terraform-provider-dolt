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

resource "dolt_database" "main" {
  name = "main"
}
