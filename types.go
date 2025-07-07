package telegraph

import (
	"fmt"
	"regexp"
	"strings"
)

// APIResponse represents the base response structure from the Telegraph API
type APIResponse struct {
	Ok     bool        `json:"ok"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// APIError represents an error response from the Telegraph API
type APIError struct {
	Code        int    `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
}

func (e *APIError) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("Telegraph API error (code %d): %s", e.Code, e.Description)
	}
	return fmt.Sprintf("Telegraph API error: %s", e.Description)
}

// Account represents a Telegraph account
type Account struct {
	ShortName  string `json:"short_name,omitempty"`
	AuthorName string `json:"author_name,omitempty"`
	AuthorURL  string `json:"author_url,omitempty"`
	// AccessToken is only returned when creating an account
	AccessToken string `json:"access_token,omitempty"`
	// AuthURL is only returned when creating an account
	AuthURL   string `json:"auth_url,omitempty"`
	PageCount int    `json:"page_count,omitempty"`
}

// Page represents a Telegraph page
type Page struct {
	Path        string `json:"path"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AuthorName  string `json:"author_name,omitempty"`
	AuthorURL   string `json:"author_url,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	Content     []Node `json:"content,omitempty"`
	Views       int    `json:"views"`
	CanEdit     bool   `json:"can_edit,omitempty"`
}

// PageList represents a list of Telegraph pages
type PageList struct {
	TotalCount int    `json:"total_count"`
	Pages      []Page `json:"pages"`
}

// PageViews represents page view statistics
type PageViews struct {
	Views int `json:"views"`
}

// Node represents a DOM node in Telegraph content
type Node struct {
	// Tag is the HTML tag name (e.g., "p", "strong", "em", "a", "br", "code", "pre", etc.)
	Tag string `json:"tag,omitempty"`
	// Attrs contains HTML attributes as key-value pairs
	Attrs map[string]string `json:"attrs,omitempty"`
	// Children contains child nodes
	Children []Node `json:"children,omitempty"`
	// Content is the text content (for text nodes)
	Content string `json:",omitempty"`
}

// CreateAccountRequest represents the request for creating a Telegraph account
type CreateAccountRequest struct {
	// ShortName is the account name (1-32 characters)
	ShortName string `json:"short_name"`
	// AuthorName is the default author name (0-128 characters)
	AuthorName string `json:"author_name,omitempty"`
	// AuthorURL is the default author URL (0-512 characters)
	AuthorURL string `json:"author_url,omitempty"`
}

// Validate validates the CreateAccountRequest
func (r *CreateAccountRequest) Validate() error {
	if r.ShortName == "" {
		return fmt.Errorf("short_name is required")
	}
	if len(r.ShortName) > 32 {
		return fmt.Errorf("short_name must be at most 32 characters")
	}
	if len(r.AuthorName) > 128 {
		return fmt.Errorf("author_name must be at most 128 characters")
	}
	if len(r.AuthorURL) > 512 {
		return fmt.Errorf("author_url must be at most 512 characters")
	}
	if r.AuthorURL != "" && !isValidURL(r.AuthorURL) {
		return fmt.Errorf("author_url must be a valid URL")
	}
	return nil
}

// EditAccountInfoRequest represents the request for editing account information
type EditAccountInfoRequest struct {
	// AccessToken is the access token of the Telegraph account
	AccessToken string `json:"access_token"`
	// ShortName is the new account name (1-32 characters)
	ShortName string `json:"short_name,omitempty"`
	// AuthorName is the new default author name (0-128 characters)
	AuthorName string `json:"author_name,omitempty"`
	// AuthorURL is the new default author URL (0-512 characters)
	AuthorURL string `json:"author_url,omitempty"`
}

// Validate validates the EditAccountInfoRequest
func (r *EditAccountInfoRequest) Validate() error {
	if r.AccessToken == "" {
		return fmt.Errorf("access_token is required")
	}
	if r.ShortName != "" && len(r.ShortName) > 32 {
		return fmt.Errorf("short_name must be at most 32 characters")
	}
	if len(r.AuthorName) > 128 {
		return fmt.Errorf("author_name must be at most 128 characters")
	}
	if len(r.AuthorURL) > 512 {
		return fmt.Errorf("author_url must be at most 512 characters")
	}
	if r.AuthorURL != "" && !isValidURL(r.AuthorURL) {
		return fmt.Errorf("author_url must be a valid URL")
	}
	return nil
}

// GetAccountInfoRequest represents the request for getting account information
type GetAccountInfoRequest struct {
	// AccessToken is the access token of the Telegraph account
	AccessToken string `json:"access_token"`
	// Fields is a list of account fields to return
	// Available fields: short_name, author_name, author_url, auth_url, page_count
	Fields []string `json:"fields,omitempty"`
}

// Validate validates the GetAccountInfoRequest
func (r *GetAccountInfoRequest) Validate() error {
	if r.AccessToken == "" {
		return fmt.Errorf("access_token is required")
	}
	
	validFields := map[string]bool{
		"short_name":  true,
		"author_name": true,
		"author_url":  true,
		"auth_url":    true,
		"page_count":  true,
	}
	
	for _, field := range r.Fields {
		if !validFields[field] {
			return fmt.Errorf("invalid field: %s", field)
		}
	}
	
	return nil
}

// CreatePageRequest represents the request for creating a Telegraph page
type CreatePageRequest struct {
	// AccessToken is the access token of the Telegraph account
	AccessToken string `json:"access_token"`
	// Title is the page title (1-256 characters)
	Title string `json:"title"`
	// AuthorName is the author name (0-128 characters)
	AuthorName string `json:"author_name,omitempty"`
	// AuthorURL is the author URL (0-512 characters)
	AuthorURL string `json:"author_url,omitempty"`
	// Content is the page content (up to 64KB)
	Content []Node `json:"content"`
	// ReturnContent determines whether to return the content in the response
	ReturnContent bool `json:"return_content,omitempty"`
}

// Validate validates the CreatePageRequest
func (r *CreatePageRequest) Validate() error {
	if r.AccessToken == "" {
		return fmt.Errorf("access_token is required")
	}
	if r.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(r.Title) > 256 {
		return fmt.Errorf("title must be at most 256 characters")
	}
	if len(r.AuthorName) > 128 {
		return fmt.Errorf("author_name must be at most 128 characters")
	}
	if len(r.AuthorURL) > 512 {
		return fmt.Errorf("author_url must be at most 512 characters")
	}
	if r.AuthorURL != "" && !isValidURL(r.AuthorURL) {
		return fmt.Errorf("author_url must be a valid URL")
	}
	if len(r.Content) == 0 {
		return fmt.Errorf("content is required")
	}
	return nil
}

// EditPageRequest represents the request for editing a Telegraph page
type EditPageRequest struct {
	// AccessToken is the access token of the Telegraph account
	AccessToken string `json:"access_token"`
	// Path is the path to the page
	Path string `json:"path"`
	// Title is the page title (1-256 characters)
	Title string `json:"title"`
	// AuthorName is the author name (0-128 characters)
	AuthorName string `json:"author_name,omitempty"`
	// AuthorURL is the author URL (0-512 characters)
	AuthorURL string `json:"author_url,omitempty"`
	// Content is the page content (up to 64KB)
	Content []Node `json:"content"`
	// ReturnContent determines whether to return the content in the response
	ReturnContent bool `json:"return_content,omitempty"`
}

// Validate validates the EditPageRequest
func (r *EditPageRequest) Validate() error {
	if r.AccessToken == "" {
		return fmt.Errorf("access_token is required")
	}
	if r.Path == "" {
		return fmt.Errorf("path is required")
	}
	if r.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(r.Title) > 256 {
		return fmt.Errorf("title must be at most 256 characters")
	}
	if len(r.AuthorName) > 128 {
		return fmt.Errorf("author_name must be at most 128 characters")
	}
	if len(r.AuthorURL) > 512 {
		return fmt.Errorf("author_url must be at most 512 characters")
	}
	if r.AuthorURL != "" && !isValidURL(r.AuthorURL) {
		return fmt.Errorf("author_url must be a valid URL")
	}
	if len(r.Content) == 0 {
		return fmt.Errorf("content is required")
	}
	return nil
}

// GetPageRequest represents the request for getting a Telegraph page
type GetPageRequest struct {
	// Path is the path to the page
	Path string `json:"path"`
	// ReturnContent determines whether to return the content in the response
	ReturnContent bool `json:"return_content,omitempty"`
}

// Validate validates the GetPageRequest
func (r *GetPageRequest) Validate() error {
	if r.Path == "" {
		return fmt.Errorf("path is required")
	}
	return nil
}

// GetPageListRequest represents the request for getting a list of Telegraph pages
type GetPageListRequest struct {
	// AccessToken is the access token of the Telegraph account
	AccessToken string `json:"access_token"`
	// Offset is the sequential number of the first page to be returned (default: 0)
	Offset int `json:"offset,omitempty"`
	// Limit is the number of pages to be returned (0-200, default: 50)
	Limit int `json:"limit,omitempty"`
}

// Validate validates the GetPageListRequest
func (r *GetPageListRequest) Validate() error {
	if r.AccessToken == "" {
		return fmt.Errorf("access_token is required")
	}
	if r.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	if r.Limit < 0 || r.Limit > 200 {
		return fmt.Errorf("limit must be between 0 and 200")
	}
	return nil
}

// GetViewsRequest represents the request for getting page views
type GetViewsRequest struct {
	// Path is the path to the page
	Path string `json:"path"`
	// Year is the required year (2000-2100)
	Year int `json:"year,omitempty"`
	// Month is the required month (1-12)
	Month int `json:"month,omitempty"`
	// Day is the required day (1-31)
	Day int `json:"day,omitempty"`
	// Hour is the required hour (0-24)
	Hour int `json:"hour,omitempty"`
}

// Validate validates the GetViewsRequest
func (r *GetViewsRequest) Validate() error {
	if r.Path == "" {
		return fmt.Errorf("path is required")
	}
	if r.Year != 0 && (r.Year < 2000 || r.Year > 2100) {
		return fmt.Errorf("year must be between 2000 and 2100")
	}
	if r.Month != 0 && (r.Month < 1 || r.Month > 12) {
		return fmt.Errorf("month must be between 1 and 12")
	}
	if r.Day != 0 && (r.Day < 1 || r.Day > 31) {
		return fmt.Errorf("day must be between 1 and 31")
	}
	if r.Hour != 0 && (r.Hour < 0 || r.Hour > 24) {
		return fmt.Errorf("hour must be between 0 and 24")
	}
	return nil
}

// isValidURL checks if a string is a valid URL
func isValidURL(str string) bool {
	if str == "" {
		return true
	}
	
	// Basic URL validation
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return urlRegex.MatchString(str)
}

// ContentBuilder provides a fluent interface for building Telegraph content
type ContentBuilder struct {
	nodes []Node
}

// NewContentBuilder creates a new content builder
func NewContentBuilder() *ContentBuilder {
	return &ContentBuilder{
		nodes: make([]Node, 0),
	}
}

// AddParagraph adds a paragraph to the content
func (cb *ContentBuilder) AddParagraph(text string) *ContentBuilder {
	cb.nodes = append(cb.nodes, Node{
		Tag: "p",
		Children: []Node{
			{Content: text},
		},
	})
	return cb
}

// AddHeading adds a heading to the content (h3 or h4)
func (cb *ContentBuilder) AddHeading(text string, level int) *ContentBuilder {
	tag := "h3"
	if level == 4 {
		tag = "h4"
	}
	
	cb.nodes = append(cb.nodes, Node{
		Tag: tag,
		Children: []Node{
			{Content: text},
		},
	})
	return cb
}

// AddLink adds a link to the content
func (cb *ContentBuilder) AddLink(text, url string) *ContentBuilder {
	cb.nodes = append(cb.nodes, Node{
		Tag: "p",
		Children: []Node{
			{
				Tag: "a",
				Attrs: map[string]string{
					"href": url,
				},
				Children: []Node{
					{Content: text},
				},
			},
		},
	})
	return cb
}

// AddImage adds an image to the content
func (cb *ContentBuilder) AddImage(src string) *ContentBuilder {
	cb.nodes = append(cb.nodes, Node{
		Tag: "img",
		Attrs: map[string]string{
			"src": src,
		},
	})
	return cb
}

// AddBlockquote adds a blockquote to the content
func (cb *ContentBuilder) AddBlockquote(text string) *ContentBuilder {
	cb.nodes = append(cb.nodes, Node{
		Tag: "blockquote",
		Children: []Node{
			{Content: text},
		},
	})
	return cb
}

// AddCodeBlock adds a code block to the content
func (cb *ContentBuilder) AddCodeBlock(code string) *ContentBuilder {
	cb.nodes = append(cb.nodes, Node{
		Tag: "pre",
		Children: []Node{
			{Content: code},
		},
	})
	return cb
}

// AddLineBreak adds a line break to the content
func (cb *ContentBuilder) AddLineBreak() *ContentBuilder {
	cb.nodes = append(cb.nodes, Node{
		Tag: "br",
	})
	return cb
}

// Build returns the built content
func (cb *ContentBuilder) Build() []Node {
	return cb.nodes
}

// String returns a string representation of the content
func (cb *ContentBuilder) String() string {
	var result strings.Builder
	for _, node := range cb.nodes {
		result.WriteString(nodeToString(node))
	}
	return result.String()
}

// nodeToString converts a Node to its string representation
func nodeToString(node Node) string {
	if node.Content != "" {
		return node.Content
	}
	
	var result strings.Builder
	for _, child := range node.Children {
		result.WriteString(nodeToString(child))
	}
	
	return result.String()
}