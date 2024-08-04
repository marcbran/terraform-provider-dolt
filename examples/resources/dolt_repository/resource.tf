terraform {
  required_providers {
    dolt = {
      source = "marcbran/dolt"
    }
  }
}

resource "dolt_repository" "main" {
  path  = "./main"
  name  = "John Doe"
  email = "john.doe@example.com"
}
