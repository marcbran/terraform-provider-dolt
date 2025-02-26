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
