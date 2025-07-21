package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/telegraph-go/telegraph"
	"golang.org/x/time/rate"
)

func main() {
	// Create a custom HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	// Create a custom retry configuration
	retryConfig := telegraph.RetryConfig{
		MaxRetries:   5,
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}

	// Create a Telegraph client with custom configuration
	client := telegraph.NewClient(
		telegraph.WithHTTPClient(httpClient),
		telegraph.WithRateLimit(rate.Limit(5)), // 5 requests per second
		telegraph.WithRetryConfig(retryConfig),
	)

	// Create account with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	account, err := client.CreateAccount(ctx, &telegraph.CreateAccountRequest{
		ShortName:  "TechBlog",
		AuthorName: "Tech Writer",
		AuthorURL:  "https://techblog.example.com",
	})
	if err != nil {
		log.Fatal("Failed to create account:", err)
	}

	fmt.Printf("Account created: %s\n", account.ShortName)
	log.Println(account)

	// Create multiple pages with different content types
	pages := []struct {
		title   string
		content []telegraph.Node
	}{
		{
			title: "Getting Started with Go",
			content: telegraph.NewContentBuilder().
				AddParagraph("Go is a programming language developed by Google.").
				AddHeading("Installation", 3).
				AddParagraph("You can download Go from the official website.").
				AddCodeBlock("go version").
				Build(),
		},
		{
			title: "Advanced Go Patterns",
			content: telegraph.NewContentBuilder().
				AddParagraph("This article covers advanced Go programming patterns.").
				AddHeading("Channels", 3).
				AddParagraph("Channels are a powerful feature in Go.").
				AddCodeBlock("ch := make(chan int)").
				AddHeading("Goroutines", 3).
				AddParagraph("Goroutines enable concurrent programming.").
				AddCodeBlock("go func() { /* code */ }()").
				Build(),
		},
		{
			title: "Go Best Practices",
			content: telegraph.NewContentBuilder().
				AddParagraph("Following best practices is important for maintainable code.").
				AddBlockquote("Clean code is not written by following a set of rules. You don't become a software craftsman by learning a list of heuristics.").
				AddHeading("Error Handling", 3).
				AddParagraph("Always handle errors explicitly in Go.").
				AddCodeBlock("if err != nil {\n    return err\n}").
				Build(),
		},
	}

	// Create pages concurrently (respecting rate limits)
	for i, pageData := range pages {
		page, err := client.CreatePage(ctx, &telegraph.CreatePageRequest{
			AccessToken:   account.AccessToken,
			Title:         pageData.title,
			AuthorName:    account.AuthorName,
			AuthorURL:     account.AuthorURL,
			Content:       pageData.content,
			ReturnContent: false,
		})
		if err != nil {
			log.Printf("Failed to create page %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Created page: %s (%s)\n", page.Title, page.URL)

		// Edit the page to add more content
		updatedContent := append(pageData.content, telegraph.Node{
			Tag: "p",
			Children: []interface{}{
				telegraph.Node{
					Content: fmt.Sprintf("This article was last updated on %s.", time.Now().Format("2006-01-02")),
				},
			},
		})

		editedPage, err := client.EditPage(ctx, &telegraph.EditPageRequest{
			AccessToken:   account.AccessToken,
			Path:          page.Path,
			Title:         page.Title,
			AuthorName:    account.AuthorName,
			AuthorURL:     account.AuthorURL,
			Content:       updatedContent,
			ReturnContent: false,
		})
		if err != nil {
			log.Printf("Failed to edit page %s: %v", page.Path, err)
			continue
		}

		fmt.Printf("Updated page: %s\n", editedPage.Title)
	}

	// Get account info with specific fields
	accountInfo, err := client.GetAccountInfo(ctx, &telegraph.GetAccountInfoRequest{
		AccessToken: account.AccessToken,
		Fields:      []string{"short_name", "author_name", "page_count"},
	})
	if err != nil {
		log.Fatal("Failed to get account info:", err)
	}

	fmt.Printf("\nAccount Info:\n")
	fmt.Printf("Short Name: %s\n", accountInfo.ShortName)
	fmt.Printf("Author Name: %s\n", accountInfo.AuthorName)
	fmt.Printf("Page Count: %d\n", accountInfo.PageCount)

	// Get paginated page list
	allPages := []telegraph.Page{}
	offset := 0
	limit := 2

	for {
		pageList, err := client.GetPageList(ctx, &telegraph.GetPageListRequest{
			AccessToken: account.AccessToken,
			Offset:      offset,
			Limit:       limit,
		})
		if err != nil {
			log.Fatal("Failed to get page list:", err)
		}

		allPages = append(allPages, pageList.Pages...)

		if len(pageList.Pages) < limit {
			break
		}

		offset += limit
	}

	fmt.Printf("\nRetrieved %d pages in total:\n", len(allPages))
	for _, page := range allPages {
		fmt.Printf("- %s (Views: %d)\n", page.Title, page.Views)

		// Get detailed view statistics
		views, err := client.GetViews(ctx, &telegraph.GetViewsRequest{
			Path: page.Path,
		})
		if err != nil {
			log.Printf("Failed to get views for %s: %v", page.Path, err)
			continue
		}

		fmt.Printf("  Total views: %d\n", views.Views)
	}

	// Demonstrate error handling
	_, err = client.GetPage(ctx, &telegraph.GetPageRequest{
		Path: "non-existent-page",
	})
	if err != nil {
		fmt.Printf("\nGeneric Error: %v\n", err)
	}

	fmt.Printf("\nDemo completed successfully!\n")
}
