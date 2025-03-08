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

resource "dolt_view" "titles" {
  depends_on = [
    dolt_table.articles
  ]

  database = dolt_database.main.name

  name  = "titles"
  query = <<EOF
SELECT title FROM articles
EOF
}
