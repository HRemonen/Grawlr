# Grawlr Usage Documentation

Grawlr is a web crawling library written in Go. This document provides an overview of how to use Grawlr, including examples and configuration options.

## Examples

The `_examples` directory contains sample implementations to help you get started with Grawlr. Below are some examples:

- [Basic Example](../_examples/basic_example.go): Demonstrates a simple web crawling setup.
- [Maximum Depth Example](../_examples/maximum_depth.go): Shows how to configure and use the depth limit feature.

To run an example, navigate to the `_examples` directory and execute the file:

```bash
cd _examples
go run basic_example.go
```

## Configuration Options

Grawlr provides several configuration options to customize its behavior. These options can be set using functional options when creating a new `Harvester` instance.

| Option               | Description                                                                                     | Default Value |
|----------------------|-------------------------------------------------------------------------------------------------|---------------|
| `WithClient`         | Sets a custom `http.Client` for the harvester.                                                  | `http.DefaultClient` |
| `WithAllowedURLs`    | Specifies a list of URLs that are allowed to be fetched.                                        | `[]` (no restrictions) |
| `WithDisallowedURLs` | Specifies a list of URLs that are disallowed from being fetched.                                | `[]` (no restrictions) |
| `WithDepthLimit`     | Sets the maximum depth of links to follow. A value of `0` means no limit.                       | `0` (no limit) |
| `WithAllowRevisit`   | Allows revisiting URLs even if they have already been visited.                                  | `false` |
| `WithContext`        | Sets a custom `context.Context` for managing request lifetimes.                                 | `context.Background()` |
| `WithStore`          | Sets a custom `Storer` implementation for caching visited URLs.                                | In-memory store |
| `WithIgnoreRobots`   | Ignores `robots.txt` rules when set to `true`.                                                  | `false` |

### Example: Configuring a Harvester

Below is an example of how to configure a `Harvester` with custom options:

```go
h := grawlr.NewHarvester(
    grawlr.WithAllowedURLs([]string{"https://example.com"}),
    grawlr.WithDisallowedURLs([]string{"https://example.com/private"}),
    grawlr.WithDepthLimit(2),
    grawlr.WithAllowRevisit(false),
    grawlr.WithIgnoreRobots(true),
)
```

## Additional Resources

For more details, refer to the [README.md](../README.md) file or explore the source code in this repository.
