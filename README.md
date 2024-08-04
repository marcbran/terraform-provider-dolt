# Terraform Provider Dolt

This repository contains a Terraform Provider that works with Dolt to create and manage SQL databases as well as the schema and data contain in them.

There is no affiliation between this repository and the official Dolt project.
This is just a personal project to learn more about Dolt and the (new) Terraform Plugin Framework.
I wouldn't recommend using it in production since I cannot promise a stable interface going forward (i.e. things might change and break your configuration).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21
- Dolt

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

Take a look at the [examples](./examples) on how to use this provider.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) and Dolt installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
