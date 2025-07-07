package examples

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telegraph-go/telegraph"
)

// Integration tests require a real Telegraph API connection
// These tests are skipped by default unless TELEGRAPH_INTEGRATION_TEST=1 is set

func TestIntegration(t *testing.T) {
	if os.Getenv("TELEGRAPH_INTEGRATION_TEST") != "1" {
		t.Skip("Integration tests skipped. Set TELEGRAPH_INTEGRATION_TEST=1 to run.")
	}

	client := telegraph.NewClient()
	ctx := context.Background()

	// Test account creation
	account, err := client.CreateAccount(ctx, &telegraph.CreateAccountRequest{
		ShortName:  "TestBlog",
		AuthorName: "Test Author",
		AuthorURL:  "https://example.com",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, account.AccessToken)
	assert.Equal(t, "TestBlog", account.ShortName)

	// Test page creation
	content := telegraph.NewContentBuilder().
		AddParagraph("This is a test article.").
		AddHeading("Test Section", 3).
		AddParagraph("This is the content of the test section.").
		Build()

	page, err := client.CreatePage(ctx, &telegraph.CreatePageRequest{
		AccessToken:   account.AccessToken,
		Title:         "Test Article",
		AuthorName:    "Test Author",
		Content:       content,
		ReturnContent: true,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, page.Path)
	assert.Equal(t, "Test Article", page.Title)
	assert.NotEmpty(t, page.URL)

	// Test page retrieval
	retrievedPage, err := client.GetPage(ctx, &telegraph.GetPageRequest{
		Path:          page.Path,
		ReturnContent: true,
	})
	require.NoError(t, err)
	assert.Equal(t, page.Title, retrievedPage.Title)
	assert.Equal(t, page.Path, retrievedPage.Path)

	// Test page editing
	updatedContent := append(content, telegraph.Node{
		Tag: "p",
		Children: []telegraph.Node{
			{Content: "This is an updated paragraph."},
		},
	})

	editedPage, err := client.EditPage(ctx, &telegraph.EditPageRequest{
		AccessToken:   account.AccessToken,
		Path:          page.Path,
		Title:         "Updated Test Article",
		Content:       updatedContent,
		ReturnContent: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Test Article", editedPage.Title)

	// Test account info retrieval
	accountInfo, err := client.GetAccountInfo(ctx, &telegraph.GetAccountInfoRequest{
		AccessToken: account.AccessToken,
		Fields:      []string{"short_name", "author_name", "page_count"},
	})
	require.NoError(t, err)
	assert.Equal(t, account.ShortName, accountInfo.ShortName)
	assert.True(t, accountInfo.PageCount > 0)

	// Test page list retrieval
	pageList, err := client.GetPageList(ctx, &telegraph.GetPageListRequest{
		AccessToken: account.AccessToken,
		Offset:      0,
		Limit:       10,
	})
	require.NoError(t, err)
	assert.True(t, pageList.TotalCount > 0)
	assert.True(t, len(pageList.Pages) > 0)

	// Test view statistics
	// Wait a bit to ensure the page has been indexed
	time.Sleep(2 * time.Second)
	
	views, err := client.GetViews(ctx, &telegraph.GetViewsRequest{
		Path: page.Path,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, views.Views, 0)

	// Test account info editing
	editedAccount, err := client.EditAccountInfo(ctx, &telegraph.EditAccountInfoRequest{
		AccessToken: account.AccessToken,
		ShortName:   "UpdatedTestBlog",
		AuthorName:  "Updated Test Author",
	})
	require.NoError(t, err)
	assert.Equal(t, "UpdatedTestBlog", editedAccount.ShortName)
	assert.Equal(t, "Updated Test Author", editedAccount.AuthorName)
}

func TestIntegrationErrorHandling(t *testing.T) {
	if os.Getenv("TELEGRAPH_INTEGRATION_TEST") != "1" {
		t.Skip("Integration tests skipped. Set TELEGRAPH_INTEGRATION_TEST=1 to run.")
	}

	client := telegraph.NewClient()
	ctx := context.Background()

	// Test invalid access token
	_, err := client.GetAccountInfo(ctx, &telegraph.GetAccountInfoRequest{
		AccessToken: "invalid-token",
		Fields:      []string{"short_name"},
	})
	require.Error(t, err)

	var apiErr *telegraph.APIError
	assert.ErrorAs(t, err, &apiErr)

	// Test non-existent page
	_, err = client.GetPage(ctx, &telegraph.GetPageRequest{
		Path: "non-existent-page-12345",
	})
	require.Error(t, err)
}

func TestIntegrationRateLimiting(t *testing.T) {
	if os.Getenv("TELEGRAPH_INTEGRATION_TEST") != "1" {
		t.Skip("Integration tests skipped. Set TELEGRAPH_INTEGRATION_TEST=1 to run.")
	}

	// Create a client with aggressive rate limiting
	client := telegraph.NewClient(
		telegraph.WithRateLimit(1), // 1 request per second
	)
	ctx := context.Background()

	start := time.Now()

	// Make multiple requests that should trigger rate limiting
	for i := 0; i < 3; i++ {
		_, err := client.GetPage(ctx, &telegraph.GetPageRequest{
			Path: "Sample-Page-12-15", // This is a sample page that should exist
		})
		// Don't assert on error as the page might not exist
		// We're just testing rate limiting behavior
		_ = err
	}

	duration := time.Since(start)
	// Should take at least 2 seconds for 3 requests with 1 RPS limit
	assert.True(t, duration >= 2*time.Second, "Rate limiting should enforce delays")
}