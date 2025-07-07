# Telegraph Go SDK

A comprehensive Go SDK for the [Telegraph API](https://telegra.ph/api) that provides a complete, type-safe interface to all Telegraph endpoints with proper error handling, rate limiting, and retry mechanisms.

[![Go Reference](https://pkg.go.dev/badge/github.com/telegraph-go/telegraph.svg)](https://pkg.go.dev/github.com/telegraph-go/telegraph)
[![Go Report Card](https://goreportcard.com/badge/github.com/telegraph-go/telegraph)](https://goreportcard.com/report/github.com/telegraph-go/telegraph)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- **Complete API Coverage**: All Telegraph API endpoints implemented
- **Type Safety**: Full type-safe request/response structs
- **Error Handling**: Comprehensive error types and handling
- **Rate Limiting**: Built-in rate limiting with configurable limits
- **Retry Logic**: Automatic retry with exponential backoff
- **Context Support**: Full context support for cancellation and timeouts
- **Thread Safe**: Concurrent usage safe
- **Configurable**: Customizable HTTP client, timeouts, and retry behavior
- **Content Builder**: Fluent interface for building Telegraph content
- **Comprehensive Testing**: >95% test coverage with unit and integration tests

## Installation

```bash
go get github.com/telegraph-go/telegraph
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/telegraph-go/telegraph"
)

func main() {
    // Create a new Telegraph client
    client := telegraph.NewClient()

    // Create a new account
    account, err := client.CreateAccount(context.Background(), &telegraph.CreateAccountRequest{
        ShortName:  "MyBlog",
        AuthorName: "John Doe",
        AuthorURL:  "https://example.com",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Account created: %s\n", account.ShortName)
    fmt.Printf("Access Token: %s\n", account.AccessToken)

    // Create a page using the content builder
    content := telegraph.NewContentBuilder().
        AddParagraph("Welcome to my first Telegraph article!").
        AddHeading("Introduction", 3).
        AddParagraph("This is a sample article.").
        AddLink("Visit our website", "https://example.com").
        Build()

    page, err := client.CreatePage(context.Background(), &telegraph.CreatePageRequest{
        AccessToken: account.AccessToken,
        Title:       "My First Article",
        Content:     content,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Page created: %s\n", page.URL)
}
```

## Configuration

The SDK provides several configuration options:

```go
import (
    "net/http"
    "time"
    "golang.org/x/time/rate"
)

// Create a custom HTTP client
httpClient := &http.Client{
    Timeout: 30 * time.Second,
}

// Custom retry configuration
retryConfig := telegraph.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 100 * time.Millisecond,
    MaxDelay:     10 * time.Second,
    Multiplier:   2.0,
}

// Create client with custom options
client := telegraph.NewClient(
    telegraph.WithHTTPClient(httpClient),
    telegraph.WithBaseURL("https://api.telegra.ph"),
    telegraph.WithRateLimit(rate.Limit(10)), // 10 requests per second
    telegraph.WithRetryConfig(retryConfig),
)
```

## API Reference

### Account Management

#### Create Account

```go
account, err := client.CreateAccount(ctx, &telegraph.CreateAccountRequest{
    ShortName:  "MyBlog",         // Required: 1-32 characters
    AuthorName: "John Doe",       // Optional: 0-128 characters
    AuthorURL:  "https://example.com", // Optional: 0-512 characters
})
```

#### Edit Account Info

```go
account, err := client.EditAccountInfo(ctx, &telegraph.EditAccountInfoRequest{
    AccessToken: "your-access-token",
    ShortName:   "UpdatedBlog",
    AuthorName:  "Jane Doe",
})
```

#### Get Account Info

```go
account, err := client.GetAccountInfo(ctx, &telegraph.GetAccountInfoRequest{
    AccessToken: "your-access-token",
    Fields:      []string{"short_name", "author_name", "page_count"},
})
```

### Page Management

#### Create Page

```go
page, err := client.CreatePage(ctx, &telegraph.CreatePageRequest{
    AccessToken:   "your-access-token",
    Title:         "My Article",
    AuthorName:    "John Doe",
    Content:       content, // []telegraph.Node
    ReturnContent: true,
})
```

#### Edit Page

```go
page, err := client.EditPage(ctx, &telegraph.EditPageRequest{
    AccessToken: "your-access-token",
    Path:        "My-Article-12-15",
    Title:       "Updated Title",
    Content:     updatedContent,
})
```

#### Get Page

```go
page, err := client.GetPage(ctx, &telegraph.GetPageRequest{
    Path:          "My-Article-12-15",
    ReturnContent: true,
})
```

#### Get Page List

```go
pageList, err := client.GetPageList(ctx, &telegraph.GetPageListRequest{
    AccessToken: "your-access-token",
    Offset:      0,
    Limit:       50,
})
```

#### Get Page Views

```go
views, err := client.GetViews(ctx, &telegraph.GetViewsRequest{
    Path:  "My-Article-12-15",
    Year:  2023,
    Month: 12,
    Day:   15,
    Hour:  10,
})
```

## Content Builder

The SDK provides a fluent interface for building Telegraph content:

```go
content := telegraph.NewContentBuilder().
    AddParagraph("Introduction paragraph").
    AddHeading("Section Title", 3).
    AddParagraph("Section content").
    AddLink("Example Link", "https://example.com").
    AddImage("https://example.com/image.jpg").
    AddBlockquote("Important quote").
    AddCodeBlock("fmt.Println(\"Hello, World!\")").
    AddLineBreak().
    Build()
```

### Supported Content Elements

- **Paragraphs**: `AddParagraph(text)`
- **Headings**: `AddHeading(text, level)` (levels 3-4)
- **Links**: `AddLink(text, url)`
- **Images**: `AddImage(src)`
- **Blockquotes**: `AddBlockquote(text)`
- **Code Blocks**: `AddCodeBlock(code)`
- **Line Breaks**: `AddLineBreak()`

## Error Handling

The SDK provides comprehensive error handling:

```go
page, err := client.GetPage(ctx, &telegraph.GetPageRequest{
    Path: "non-existent-page",
})
if err != nil {
    var apiErr *telegraph.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error (code %d): %s\n", apiErr.Code, apiErr.Description)
    } else {
        fmt.Printf("Generic Error: %v\n", err)
    }
}
```

## Rate Limiting

The SDK includes built-in rate limiting to respect API limits:

```go
// Set custom rate limit (requests per second)
client := telegraph.NewClient(
    telegraph.WithRateLimit(rate.Limit(5)), // 5 requests per second
)
```

## Retry Logic

Automatic retry with exponential backoff for failed requests:

```go
retryConfig := telegraph.RetryConfig{
    MaxRetries:   3,                        // Maximum retry attempts
    InitialDelay: 100 * time.Millisecond,   // Initial delay
    MaxDelay:     5 * time.Second,          // Maximum delay
    Multiplier:   2.0,                      // Backoff multiplier
}

client := telegraph.NewClient(
    telegraph.WithRetryConfig(retryConfig),
)
```

## Testing

Run the unit tests:

```bash
go test ./...
```

Run integration tests (requires internet connection):

```bash
TELEGRAPH_INTEGRATION_TEST=1 go test ./...
```

Run benchmarks:

```bash
go test -bench=. ./...
```

## Examples

Check out the [examples](./examples) directory for more comprehensive usage examples:

- [Basic Usage](./examples/basic/main.go) - Simple account and page creation
- [Advanced Usage](./examples/advanced/main.go) - Advanced features and configuration
- [Integration Tests](./examples/integration_test.go) - Real API integration tests

## Thread Safety

The Telegraph client is thread-safe and can be used concurrently from multiple goroutines:

```go
client := telegraph.NewClient()

// Safe to use from multiple goroutines
for i := 0; i < 10; i++ {
    go func(i int) {
        page, err := client.GetPage(ctx, &telegraph.GetPageRequest{
            Path: fmt.Sprintf("page-%d", i),
        })
        // Handle page and err
    }(i)
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

### v1.0.0

- Initial release
- Complete Telegraph API coverage
- Type-safe request/response structs
- Comprehensive error handling
- Rate limiting and retry logic
- Content builder utility
- Full test coverage
- Documentation and examples

## Support

If you have questions or need help, please:

1. Check the [documentation](https://pkg.go.dev/github.com/telegraph-go/telegraph)
2. Look at the [examples](./examples)
3. Open an issue on GitHub

## Acknowledgments

- [Telegraph API](https://telegra.ph/api) for providing the excellent API
- The Go community for inspiration and best practices