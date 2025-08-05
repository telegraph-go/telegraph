package telegraph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestNewClient(t *testing.T) {
	t.Run("default client", func(t *testing.T) {
		client := NewClient()
		assert.NotNil(t, client)
		assert.Equal(t, "https://api.telegra.ph", client.baseURL)
		assert.NotNil(t, client.httpClient)
		assert.NotNil(t, client.rateLimiter)
		assert.Equal(t, DefaultRetryConfig, client.retryConfig)
	})

	t.Run("with custom options", func(t *testing.T) {
		httpClient := &http.Client{Timeout: 10 * time.Second}
		retryConfig := RetryConfig{MaxRetries: 5}

		client := NewClient(
			WithHTTPClient(httpClient),
			WithBaseURL("https://custom.api.com"),
			WithRateLimit(rate.Limit(5)),
			WithRetryConfig(retryConfig),
		)

		assert.Equal(t, httpClient, client.httpClient)
		assert.Equal(t, "https://custom.api.com", client.baseURL)
		assert.Equal(t, retryConfig, client.retryConfig)
	})
}

func TestClientCreateAccount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/createAccount", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req CreateAccountRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "TestBlog", req.ShortName)
		assert.Equal(t, "John Doe", req.AuthorName)
		assert.Equal(t, "https://example.com", req.AuthorURL)

		resp := APIResponse{
			Ok: true,
			Result: Account{
				ShortName:   req.ShortName,
				AuthorName:  req.AuthorName,
				AuthorURL:   req.AuthorURL,
				AccessToken: "test-access-token",
				AuthURL:     "https://edit.telegra.ph/auth/test-auth-url",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	account, err := client.CreateAccount(context.Background(), &CreateAccountRequest{
		ShortName:  "TestBlog",
		AuthorName: "John Doe",
		AuthorURL:  "https://example.com",
	})

	require.NoError(t, err)
	assert.Equal(t, "TestBlog", account.ShortName)
	assert.Equal(t, "John Doe", account.AuthorName)
	assert.Equal(t, "https://example.com", account.AuthorURL)
	assert.Equal(t, "test-access-token", account.AccessToken)
	assert.Equal(t, "https://edit.telegra.ph/auth/test-auth-url", account.AuthURL)
}

func TestClientCreateAccountValidation(t *testing.T) {
	client := NewClient()

	t.Run("missing short name", func(t *testing.T) {
		_, err := client.CreateAccount(context.Background(), &CreateAccountRequest{
			AuthorName: "John Doe",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "short_name is required")
	})

	t.Run("short name too long", func(t *testing.T) {
		_, err := client.CreateAccount(context.Background(), &CreateAccountRequest{
			ShortName: strings.Repeat("a", 33),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "short_name must be at most 32 characters")
	})
}

func TestClientCreatePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/createPage", r.URL.Path)

		var req CreatePageRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "test-token", req.AccessToken)
		assert.Equal(t, "Test Article", req.Title)
		assert.Len(t, req.Content, 1)
		assert.Equal(t, "p", req.Content[0].Tag)

		resp := APIResponse{
			Ok: true,
			Result: Page{
				Path:        "Test-Article-12-15",
				URL:         "https://telegra.ph/Test-Article-12-15",
				Title:       req.Title,
				Description: "Test description",
				Views:       0,
				CanEdit:     true,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	page, err := client.CreatePage(context.Background(), &CreatePageRequest{
		AccessToken: "test-token",
		Title:       "Test Article",
		Content: []Node{
			{
				Tag: "p",
				Children: []interface{}{
					Node{Content: "Hello, World!"},
				},
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "Test-Article-12-15", page.Path)
	assert.Equal(t, "https://telegra.ph/Test-Article-12-15", page.URL)
	assert.Equal(t, "Test Article", page.Title)
	assert.True(t, page.CanEdit)
}

func TestClientGetPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/getPage", r.URL.Path)
		assert.Equal(t, "Test-Article-12-15", r.URL.Query().Get("path"))
		assert.Equal(t, "true", r.URL.Query().Get("return_content"))

		resp := APIResponse{
			Ok: true,
			Result: Page{
				Path:        "Test-Article-12-15",
				URL:         "https://telegra.ph/Test-Article-12-15",
				Title:       "Test Article",
				Description: "Test description",
				Views:       42,
				Content: []Node{
					{
						Tag: "p",
						Children: []interface{}{
							Node{Content: "Hello, World!"},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	page, err := client.GetPage(context.Background(), &GetPageRequest{
		Path:          "Test-Article-12-15",
		ReturnContent: true,
	})

	require.NoError(t, err)
	assert.Equal(t, "Test-Article-12-15", page.Path)
	assert.Equal(t, "Test Article", page.Title)
	assert.Equal(t, 42, page.Views)
	assert.Len(t, page.Content, 1)
}

func TestClientGetPageList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/getPageList", r.URL.Path)

		var req GetPageListRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "test-token", req.AccessToken)
		assert.Equal(t, 0, req.Offset)
		assert.Equal(t, 10, req.Limit)

		resp := APIResponse{
			Ok: true,
			Result: PageList{
				TotalCount: 1,
				Pages: []Page{
					{
						Path:        "Test-Article-12-15",
						URL:         "https://telegra.ph/Test-Article-12-15",
						Title:       "Test Article",
						Description: "Test description",
						Views:       42,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	pageList, err := client.GetPageList(context.Background(), &GetPageListRequest{
		AccessToken: "test-token",
		Offset:      0,
		Limit:       10,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, pageList.TotalCount)
	assert.Len(t, pageList.Pages, 1)
	assert.Equal(t, "Test-Article-12-15", pageList.Pages[0].Path)
}

func TestClientGetViews(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/getViews", r.URL.Path)

		var req GetViewsRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "Test-Article-12-15", req.Path)
		assert.Equal(t, 2023, req.Year)
		assert.Equal(t, 12, req.Month)

		resp := APIResponse{
			Ok: true,
			Result: PageViews{
				Views: 100,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	views, err := client.GetViews(context.Background(), &GetViewsRequest{
		Path:  "Test-Article-12-15",
		Year:  2023,
		Month: 12,
	})

	require.NoError(t, err)
	assert.Equal(t, 100, views.Views)
}

func TestClientErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := APIError{
			Code:        400,
			Description: "Bad Request",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	_, err := client.CreateAccount(context.Background(), &CreateAccountRequest{
		ShortName: "Test",
	})

	require.Error(t, err)
	var apiErr *APIError
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.Code)
	assert.Equal(t, "Bad Request", apiErr.Description)
}

func TestClientRetryLogic(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp := APIResponse{
			Ok: true,
			Result: Account{
				ShortName:   "Test",
				AccessToken: "test-token",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRetryConfig(RetryConfig{
			MaxRetries:   3,
			InitialDelay: 1 * time.Millisecond,
			MaxDelay:     10 * time.Millisecond,
			Multiplier:   2.0,
		}),
	)

	account, err := client.CreateAccount(context.Background(), &CreateAccountRequest{
		ShortName: "Test",
	})

	require.NoError(t, err)
	assert.Equal(t, "Test", account.ShortName)
	assert.Equal(t, 3, attempts)
}

func TestClientRateLimiting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{
			Ok: true,
			Result: Account{
				ShortName:   "Test",
				AccessToken: "test-token",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRateLimit(rate.Limit(1)), // 1 request per second
	)

	start := time.Now()

	// Make two requests
	for i := 0; i < 2; i++ {
		_, err := client.CreateAccount(context.Background(), &CreateAccountRequest{
			ShortName: fmt.Sprintf("Test%d", i),
		})
		require.NoError(t, err)
	}

	duration := time.Since(start)
	// Should take at least 1 second due to rate limiting
	assert.True(t, duration >= 1*time.Second)
}

func TestClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		resp := APIResponse{
			Ok: true,
			Result: Account{
				ShortName:   "Test",
				AccessToken: "test-token",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.CreateAccount(ctx, &CreateAccountRequest{
		ShortName: "Test",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestConvertHTMLToPage(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name         string
		html         string
		opts         *HTMLToPageOptions
		expectedPage *Page
		expectedErr  error
	}{
		{
			name: "simple html",
			html: `<html><head><title>My Title</title></head><body><p>Hello, <b>world</b>!</p></body></html>`,
			expectedPage: &Page{
				Title: "My Title",
				Content: []Node{
					{Tag: "p", Children: []interface{}{"Hello, ", Node{Tag: "strong", Children: []interface{}{"world"}}, "!"}},
				},
			},
		},
		{
			name: "html with meta tags",
			html: `<html><head><title>Test Article</title><meta name="author" content="Jane Doe"><meta name="url" content="https://example.com/jane"></head><body><h1>Welcome</h1><p>Content here.</p></body></html>`,
			expectedPage: &Page{
				Title:      "Test Article",
				AuthorName: "Jane Doe",
				AuthorURL:  "https://example.com/jane",
				Content: []Node{
					{Tag: "h3", Children: []interface{}{"Welcome"}},
					{Tag: "p", Children: []interface{}{"Content here."}},
				},
			},
		},
		{
			name: "html with options overriding meta tags",
			html: `<html><head><title>Test Article</title><meta name="author" content="Jane Doe"></head><body><p>Content here.</p></body></html>`,
			opts: &HTMLToPageOptions{AuthorName: "Override Author"},
			expectedPage: &Page{
				Title:      "Test Article",
				AuthorName: "Override Author",
				Content: []Node{
					{Tag: "p", Children: []interface{}{"Content here."}},
				},
			},
		},
		{
			name: "html with unsupported tags and scripts",
			html: `<html><head><title>Mixed Content</title></head><body><section><h2>Section Title</h2><p>Some text.</p><script>alert('hi');</script><div>Another div</div><ul><li>Item 1</li></ul></body></html>`,
			expectedPage: &Page{
				Title: "Mixed Content",
				Content: []Node{
					{Tag: "p", Children: []interface{}{
						Node{Tag: "h3", Children: []interface{}{"Section Title"}},
						Node{Tag: "p", Children: []interface{}{"Some text."}},
						Node{Tag: "p", Children: []interface{}{"Another div"}},
						Node{Tag: "ul", Children: []interface{}{Node{Tag: "li", Children: []interface{}{"Item 1"}}}},
					}},
				},
			},
		},
		{
			name: "html with image and link",
			html: `<html><body><p>Check this <a href="https://example.com">link</a> and an <img src="image.jpg"> image.</p></body></html>`,
			expectedPage: &Page{
				Content: []Node{
					{Tag: "p", Children: []interface{}{"Check this ", Node{Tag: "a", Attrs: map[string]string{"href": "https://example.com"}, Children: []interface{}{"link"}}, " and an ", Node{Tag: "img", Attrs: map[string]string{"src": "image.jpg"}}, " image."}},
				},
			},
		},
		{
			name: "empty body",
			html: `<html><head><title>Empty</title></head><body></body></html>`,
			expectedPage: &Page{
				Title:   "Empty",
				Content: []Node{},
			},
		},
		{
			name: "no body tag",
			html: `<html><head><title>No Body</title></head></html>`,
			expectedPage: &Page{
				Title:   "No Body",
				Content: []Node{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, err := client.ConvertHTMLToPage(tt.html, tt.opts)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedPage.Title, page.Title)
			assert.Equal(t, tt.expectedPage.AuthorName, page.AuthorName)
			assert.Equal(t, tt.expectedPage.AuthorURL, page.AuthorURL)
			assert.Equal(t, tt.expectedPage.Description, page.Description)

			// Custom assertion for content due to interface{} slice comparison complexities
			assertNodesEqual(t, tt.expectedPage.Content, page.Content)
		})
	}
}

// assertNodesEqual recursively compares two slices of Node objects
func assertNodesEqual(t *testing.T, expected, actual []Node) bool {
	if !assert.Len(t, actual, len(expected), "Node slices should have the same length") {
		return false
	}

	for i := range expected {
		if !assertNodeEqual(t, expected[i], actual[i]) {
			return false
		}
	}
	return true
}

// assertNodeEqual recursively compares two Node objects
func assertNodeEqual(t *testing.T, expected, actual Node) bool {
	if !assert.Equal(t, expected.Tag, actual.Tag, "Tags should be equal") {
		return false
	}
	if !assert.Equal(t, expected.Content, actual.Content, "Contents should be equal") {
		return false
	}

	if !assert.Equal(t, expected.Attrs, actual.Attrs, "Attributes should be equal") {
		return false
	}

	// Compare children recursively
	if !assert.Len(t, actual.Children, len(expected.Children), "Children slices should have the same length") {
		return false
	}
	for i := range expected.Children {
		expChild := expected.Children[i]
		actChild := actual.Children[i]

		switch expC := expChild.(type) {
		case string:
			if !assert.IsType(t, expC, actChild, "Child should be string") {
				return false
			}
			if !assert.Equal(t, expC, actChild, "String children should be equal") {
				return false
			}
		case Node:
			if !assert.IsType(t, expC, actChild, "Child should be Node") {
				return false
			}
			if !assertNodeEqual(t, expC, actChild.(Node)) {
				return false
			}
		default:
			return assert.Fail(t, "Unexpected type in expected children")
		}
	}
	return true
}

// Benchmark tests
func BenchmarkClientCreateAccount(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{
			Ok: true,
			Result: Account{
				ShortName:   "Test",
				AccessToken: "test-token",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.CreateAccount(context.Background(), &CreateAccountRequest{
			ShortName: "Test",
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClientCreatePage(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{
			Ok: true,
			Result: Page{
				Path:  "Test-Article-12-15",
				URL:   "https://telegra.ph/Test-Article-12-15",
				Title: "Test Article",
				Views: 0,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.CreatePage(context.Background(), &CreatePageRequest{
			AccessToken: "test-token",
			Title:       "Test Article",
			Content: []Node{
				{
					Tag: "p",
					Children: []interface{}{
						Node{Content: "Hello, World!"},
					},
				},
			},
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}
