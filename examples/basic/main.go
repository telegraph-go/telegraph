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
		log.Fatal("Failed to create account:", err)
	}

	fmt.Printf("Account created successfully!\n")
	fmt.Printf("Short Name: %s\n", account.ShortName)
	fmt.Printf("Author Name: %s\n", account.AuthorName)
	fmt.Printf("Access Token: %s\n", account.AccessToken)
	fmt.Printf("Auth URL: %s\n", account.AuthURL)

	// Create a simple page using the content builder
	content := telegraph.NewContentBuilder().
		AddParagraph("Welcome to my first Telegraph article!").
		AddHeading("Introduction", 3).
		AddParagraph("This is a sample article created using the Telegraph Go SDK.").
		AddLink("Visit our website", "https://example.com").
		AddLineBreak().
		AddBlockquote("Telegraph is a minimalist publishing tool that allows you to create richly formatted posts and push them to the web in just a click.").
		AddCodeBlock("fmt.Println(\"Hello, Telegraph!\")").
		Build()

	page, err := client.CreatePage(context.Background(), &telegraph.CreatePageRequest{
		AccessToken:   account.AccessToken,
		Title:         "My First Telegraph Article",
		AuthorName:    "John Doe",
		AuthorURL:     "https://example.com",
		Content:       content,
		ReturnContent: true,
	})
	if err != nil {
		log.Fatal("Failed to create page:", err)
	}

	fmt.Printf("\nPage created successfully!\n")
	fmt.Printf("Path: %s\n", page.Path)
	fmt.Printf("URL: %s\n", page.URL)
	fmt.Printf("Title: %s\n", page.Title)
	fmt.Printf("Description: %s\n", page.Description)
	fmt.Printf("Views: %d\n", page.Views)

	// Get the page we just created
	retrievedPage, err := client.GetPage(context.Background(), &telegraph.GetPageRequest{
		Path:          page.Path,
		ReturnContent: true,
	})
	if err != nil {
		log.Fatal("Failed to get page:", err)
	}

	fmt.Printf("\nRetrieved page successfully!\n")
	fmt.Printf("Title: %s\n", retrievedPage.Title)
	fmt.Printf("Views: %d\n", retrievedPage.Views)
	fmt.Printf("Content nodes: %d\n", len(retrievedPage.Content))

	// Get page list
	pageList, err := client.GetPageList(context.Background(), &telegraph.GetPageListRequest{
		AccessToken: account.AccessToken,
		Offset:      0,
		Limit:       10,
	})
	if err != nil {
		log.Fatal("Failed to get page list:", err)
	}

	fmt.Printf("\nPage list retrieved successfully!\n")
	fmt.Printf("Total pages: %d\n", pageList.TotalCount)
	fmt.Printf("Pages in this response: %d\n", len(pageList.Pages))

	// Get view statistics
	views, err := client.GetViews(context.Background(), &telegraph.GetViewsRequest{
		Path: page.Path,
	})
	if err != nil {
		log.Fatal("Failed to get views:", err)
	}

	fmt.Printf("\nView statistics retrieved successfully!\n")
	fmt.Printf("Total views: %d\n", views.Views)
}