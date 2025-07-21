// Package telegraph provides a comprehensive Go SDK for the Telegraph API (https://telegra.ph/api)
//
// The Telegraph API is a publishing platform that allows you to create and manage
// articles on telegra.ph. This SDK provides a complete, type-safe interface to all
// Telegraph API endpoints with proper error handling, rate limiting, and retry mechanisms.
//
// Basic usage:
//
//	client := telegraph.NewClient()
//
//	// Create an account
//	account, err := client.CreateAccount(context.Background(), &telegraph.CreateAccountRequest{
//		ShortName: "MyBlog",
//		AuthorName: "John Doe",
//		AuthorURL: "https://example.com",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create a page
//	page, err := client.CreatePage(context.Background(), &telegraph.CreatePageRequest{
//		AccessToken: account.AccessToken,
//		Title: "My First Article",
//		Content: []telegraph.Node{
//			{Tag: "p", Children: []telegraph.Node{{Content: "Hello, World!"}}},
//		},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
package telegraph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Client represents the Telegraph API client
type Client struct {
	httpClient  *http.Client
	baseURL     string
	rateLimiter *rate.Limiter
	retryConfig RetryConfig
	mu          sync.RWMutex
}

// RetryConfig defines retry behavior for failed requests
type RetryConfig struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultRetryConfig provides sensible defaults for retry behavior
var DefaultRetryConfig = RetryConfig{
	MaxRetries:   3,
	InitialDelay: 100 * time.Millisecond,
	MaxDelay:     5 * time.Second,
	Multiplier:   2.0,
}

// ClientOption represents a configuration option for the Telegraph client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithBaseURL sets a custom base URL for the API
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = strings.TrimSuffix(baseURL, "/")
	}
}

// WithRateLimit sets the rate limit for API requests (requests per second)
func WithRateLimit(rps rate.Limit) ClientOption {
	return func(c *Client) {
		c.rateLimiter = rate.NewLimiter(rps, int(rps))
	}
}

// WithRetryConfig sets the retry configuration
func WithRetryConfig(config RetryConfig) ClientOption {
	return func(c *Client) {
		c.retryConfig = config
	}
}

// NewClient creates a new Telegraph API client with the provided options
func NewClient(opts ...ClientOption) *Client {
	client := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:     "https://api.telegra.ph",
		rateLimiter: rate.NewLimiter(rate.Limit(10), 10), // 10 requests per second by default
		retryConfig: DefaultRetryConfig,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// doRequest performs an HTTP request with retry logic and rate limiting
func (c *Client) doRequest(ctx context.Context, method, endpoint string, data interface{}) (*http.Response, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Apply rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiting failed: %w", err)
	}

	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	url := fmt.Sprintf("%s/%s", c.baseURL, strings.TrimPrefix(endpoint, "/"))

	var lastErr error
	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := c.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "telegraph-go-sdk/1.0.0")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if !c.shouldRetry(err) {
				return nil, fmt.Errorf("request failed: %w", err)
			}
			continue
		}

		// Check if we should retry based on status code
		if c.shouldRetryStatus(resp.StatusCode) {
			resp.Body.Close()
			lastErr = fmt.Errorf("received status code %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.retryConfig.MaxRetries+1, lastErr)
}

func (c *Client) calculateDelay(attempt int) time.Duration {
	delay := c.retryConfig.InitialDelay * time.Duration(1<<uint(attempt-1)) * time.Duration(c.retryConfig.Multiplier)

	if delay > c.retryConfig.MaxDelay {
		delay = c.retryConfig.MaxDelay
	}

	return delay
}

// shouldRetry determines if a request should be retried based on the error
func (c *Client) shouldRetry(err error) bool {
	// Retry on network errors, timeouts, etc.
	return true
}

// shouldRetryStatus determines if a request should be retried based on status code
func (c *Client) shouldRetryStatus(statusCode int) bool {
	// Retry on 5xx errors and 429 (Too Many Requests)
	return statusCode >= 500 || statusCode == 429
}

// parseResponse parses the API response and handles errors
func (c *Client) parseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return &APIError{
				Code:        resp.StatusCode,
				Description: string(body),
			}
		}
		return &apiErr
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Ok {
		return &APIError{
			Code:        0,
			Description: "API returned ok: false",
		}
	}

	if result != nil {
		resultBytes, err := json.Marshal(apiResp.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		if err := json.Unmarshal(resultBytes, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

// CreateAccount creates a new Telegraph account
//
// This method is used to create a new Telegraph account. Most users only need one account,
// but this can be useful for channel administrators who would like to keep individual
// author names and profile links for each of their channels.
//
// Example:
//
//	account, err := client.CreateAccount(ctx, &telegraph.CreateAccountRequest{
//		ShortName:  "MyBlog",
//		AuthorName: "John Doe",
//		AuthorURL:  "https://example.com",
//	})
func (c *Client) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*Account, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, "POST", "/createAccount", req)
	if err != nil {
		return nil, err
	}

	var account Account
	if err := c.parseResponse(resp, &account); err != nil {
		return nil, err
	}

	return &account, nil
}

// EditAccountInfo edits the account information
//
// This method is used to update information about a Telegraph account.
// Pass only the parameters that you want to edit.
//
// Example:
//
//	account, err := client.EditAccountInfo(ctx, &telegraph.EditAccountInfoRequest{
//		AccessToken: "your-access-token",
//		ShortName:   "UpdatedBlog",
//		AuthorName:  "Jane Doe",
//	})
func (c *Client) EditAccountInfo(ctx context.Context, req *EditAccountInfoRequest) (*Account, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, "POST", "/editAccountInfo", req)
	if err != nil {
		return nil, err
	}

	var account Account
	if err := c.parseResponse(resp, &account); err != nil {
		return nil, err
	}

	return &account, nil
}

// GetAccountInfo gets account information
//
// This method is used to get information about a Telegraph account.
// Returns an Account object on success.
//
// Example:
//
//	account, err := client.GetAccountInfo(ctx, &telegraph.GetAccountInfoRequest{
//		AccessToken: "your-access-token",
//		Fields:      []string{"short_name", "author_name", "page_count"},
//	})
func (c *Client) GetAccountInfo(ctx context.Context, req *GetAccountInfoRequest) (*Account, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, "POST", "/getAccountInfo", req)
	if err != nil {
		return nil, err
	}

	var account Account
	if err := c.parseResponse(resp, &account); err != nil {
		return nil, err
	}

	return &account, nil
}

// CreatePage creates a new Telegraph page
//
// This method is used to create a new Telegraph page. Returns a Page object on success.
//
// Example:
//
//	page, err := client.CreatePage(ctx, &telegraph.CreatePageRequest{
//		AccessToken: "your-access-token",
//		Title:       "My Article",
//		Content: []telegraph.Node{
//			{Tag: "p", Children: []telegraph.Node{{Content: "Hello, World!"}}},
//		},
//	})
func (c *Client) CreatePage(ctx context.Context, req *CreatePageRequest) (*Page, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, "POST", "/createPage", req)
	if err != nil {
		return nil, err
	}

	var page Page
	if err := c.parseResponse(resp, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// EditPage edits an existing Telegraph page
//
// This method is used to edit an existing Telegraph page. Returns a Page object on success.
//
// Example:
//
//	page, err := client.EditPage(ctx, &telegraph.EditPageRequest{
//		AccessToken: "your-access-token",
//		Path:        "My-Article-12-15",
//		Title:       "Updated Article Title",
//		Content: []telegraph.Node{
//			{Tag: "p", Children: []telegraph.Node{{Content: "Updated content"}}},
//		},
//	})
func (c *Client) EditPage(ctx context.Context, req *EditPageRequest) (*Page, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, "POST", "/editPage", req)
	if err != nil {
		return nil, err
	}

	var page Page
	if err := c.parseResponse(resp, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// GetPage gets a Telegraph page
//
// This method is used to get a Telegraph page. Returns a Page object on success.
//
// Example:
//
//	page, err := client.GetPage(ctx, &telegraph.GetPageRequest{
//		Path:         "My-Article-12-15",
//		ReturnContent: true,
//	})
func (c *Client) GetPage(ctx context.Context, req *GetPageRequest) (*Page, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// For GET requests, we need to build query parameters
	params := url.Values{}
	params.Add("path", req.Path)
	if req.ReturnContent {
		params.Add("return_content", "true")
	}

	endpoint := fmt.Sprintf("/getPage?%s", params.Encode())
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var page Page
	if err := c.parseResponse(resp, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// GetPageList gets a list of pages belonging to a Telegraph account
//
// This method is used to get a list of pages belonging to a Telegraph account.
// Returns a PageList object on success.
//
// Example:
//
//	pageList, err := client.GetPageList(ctx, &telegraph.GetPageListRequest{
//		AccessToken: "your-access-token",
//		Offset:      0,
//		Limit:       10,
//	})
func (c *Client) GetPageList(ctx context.Context, req *GetPageListRequest) (*PageList, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, "POST", "/getPageList", req)
	if err != nil {
		return nil, err
	}

	var pageList PageList
	if err := c.parseResponse(resp, &pageList); err != nil {
		return nil, err
	}

	return &pageList, nil
}

// GetViews gets the number of views for a Telegraph page
//
// This method is used to get the number of views for a Telegraph page.
// Returns a PageViews object on success.
//
// Example:
//
//	views, err := client.GetViews(ctx, &telegraph.GetViewsRequest{
//		Path: "My-Article-12-15",
//		Year: 2023,
//		Month: 12,
//		Day: 15,
//		Hour: 10,
//	})
func (c *Client) GetViews(ctx context.Context, req *GetViewsRequest) (*PageViews, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, "POST", "/getViews", req)
	if err != nil {
		return nil, err
	}

	var views PageViews
	if err := c.parseResponse(resp, &views); err != nil {
		return nil, err
	}

	return &views, nil
}
