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

resource "dolt_table" "articles" {
  repository_path = dolt_repository.main.path
  author_name     = dolt_repository.main.name
  author_email    = dolt_repository.main.email

  name  = "articles"
  query = <<EOF
CREATE TABLE articles (
  id INT PRIMARY KEY,
  title VARCHAR(128) UNIQUE,
);
EOF
}

resource "dolt_rowset" "rowset" {
  repository_path = dolt_repository.main.path
  author_name     = dolt_repository.main.name
  author_email    = dolt_repository.main.email
  table_name      = dolt_table.articles.name

  columns       = ["id", "title"]
  unique_column = "id"
  values = {
    1 = ["1", "How to use Dolt"],
    2 = ["2", "Terraform Internals"],
  }
}
