# Grawlr

**Grawlr**, a simple web crawler written in Go

## Table of Contents
- [Installation](#installation)
- [Usage](#usage)
- [Testing](#testing)
- [Linting](#linting)

## Installation

### Prerequisites
- [Go](https://golang.org/doc/install) (version 1.23+)

### Clone the Repository

To download the source code, clone the repository:

```bash
git clone git@github.com:HRemonen/Grawlr.git
cd grawlr
```

### Install Dependencies

Run the following command to install Go module dependencies:

```bash
go mod tidy
```

This will install any necessary packages for the project.

## Usage

Currently, the project does not actually expose any function or interface to use... TBD

## Testing

This project includes tests for various different modules.

To run the tests for a single package, use the command:

```bash
go test -v <package name dir>
```

To run all the tests:

```bash
go test ./...
```

## Linting

To ensure that the codebase follows Go best practices and maintain a clean, consistent style, we use `golangci-lint`, a popular linter aggregator for Go.

### Installing the Linter

First, install `golangci-lint` by following the official instructions [here](https://golangci-lint.run/usage/install/). You can also install it using `go install`:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Running the Linter

Once installed, you can run the linter on the project using the following command:

```bash
golangci-lint run
```

This will check the entire codebase for issues and display any linting errors, warnings, or suggestions.

In some cases, the linter can automatically fix issues like formatting errors. To apply fixes automatically, run:

```bash
golangci-lint run --fix
```

The linter is also run on the CI pipeline.




