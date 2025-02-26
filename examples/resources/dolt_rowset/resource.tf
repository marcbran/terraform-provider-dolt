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

resource "dolt_table" "articles" {
  database = dolt_database.main.name

  name  = "articles"
  query = <<EOF
CREATE TABLE articles (
  id INT PRIMARY KEY,
  title VARCHAR(128) UNIQUE
);
EOF
}

resource "dolt_rowset" "rowset" {
  database = dolt_database.main.name
  table    = dolt_table.articles.name

  columns       = ["id", "title"]
  unique_column = "id"
  values = {
    1 = ["1", "How to use Dolt"],
    2 = ["2", "Terraform Internals"],
  }
}
