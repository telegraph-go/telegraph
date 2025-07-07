package telegraph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAccountRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateAccountRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: CreateAccountRequest{
				ShortName:  "TestBlog",
				AuthorName: "John Doe",
				AuthorURL:  "https://example.com",
			},
			wantErr: false,
		},
		{
			name:    "missing short name",
			req:     CreateAccountRequest{},
			wantErr: true,
			errMsg:  "short_name is required",
		},
		{
			name: "short name too long",
			req: CreateAccountRequest{
				ShortName: "This is a very long short name that exceeds the maximum length of 32 characters",
			},
			wantErr: true,
			errMsg:  "short_name must be at most 32 characters",
		},
		{
			name: "author name too long",
			req: CreateAccountRequest{
				ShortName:  "Test",
				AuthorName: "This is a very long author name that definitely exceeds the maximum allowed length of 128 characters for the author name field",
			},
			wantErr: true,
			errMsg:  "author_name must be at most 128 characters",
		},
		{
			name: "invalid author URL",
			req: CreateAccountRequest{
				ShortName: "Test",
				AuthorURL: "not-a-valid-url",
			},
			wantErr: true,
			errMsg:  "author_url must be a valid URL",
		},
		{
			name: "empty author URL is valid",
			req: CreateAccountRequest{
				ShortName: "Test",
				AuthorURL: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEditAccountInfoRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     EditAccountInfoRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: EditAccountInfoRequest{
				AccessToken: "test-token",
				ShortName:   "UpdatedBlog",
				AuthorName:  "Jane Doe",
				AuthorURL:   "https://example.com",
			},
			wantErr: false,
		},
		{
			name:    "missing access token",
			req:     EditAccountInfoRequest{},
			wantErr: true,
			errMsg:  "access_token is required",
		},
		{
			name: "short name too long",
			req: EditAccountInfoRequest{
				AccessToken: "test-token",
				ShortName:   "This is a very long short name that exceeds the maximum length",
			},
			wantErr: true,
			errMsg:  "short_name must be at most 32 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetAccountInfoRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     GetAccountInfoRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: GetAccountInfoRequest{
				AccessToken: "test-token",
				Fields:      []string{"short_name", "author_name", "page_count"},
			},
			wantErr: false,
		},
		{
			name:    "missing access token",
			req:     GetAccountInfoRequest{},
			wantErr: true,
			errMsg:  "access_token is required",
		},
		{
			name: "invalid field",
			req: GetAccountInfoRequest{
				AccessToken: "test-token",
				Fields:      []string{"invalid_field"},
			},
			wantErr: true,
			errMsg:  "invalid field: invalid_field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreatePageRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreatePageRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: CreatePageRequest{
				AccessToken: "test-token",
				Title:       "Test Article",
				Content: []Node{
					{Tag: "p", Children: []Node{{Content: "Hello, World!"}}},
				},
			},
			wantErr: false,
		},
		{
			name:    "missing access token",
			req:     CreatePageRequest{},
			wantErr: true,
			errMsg:  "access_token is required",
		},
		{
			name: "missing title",
			req: CreatePageRequest{
				AccessToken: "test-token",
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "missing content",
			req: CreatePageRequest{
				AccessToken: "test-token",
				Title:       "Test Article",
			},
			wantErr: true,
			errMsg:  "content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetPageListRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     GetPageListRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: GetPageListRequest{
				AccessToken: "test-token",
				Offset:      0,
				Limit:       10,
			},
			wantErr: false,
		},
		{
			name:    "missing access token",
			req:     GetPageListRequest{},
			wantErr: true,
			errMsg:  "access_token is required",
		},
		{
			name: "negative offset",
			req: GetPageListRequest{
				AccessToken: "test-token",
				Offset:      -1,
			},
			wantErr: true,
			errMsg:  "offset must be non-negative",
		},
		{
			name: "limit too high",
			req: GetPageListRequest{
				AccessToken: "test-token",
				Limit:       201,
			},
			wantErr: true,
			errMsg:  "limit must be between 0 and 200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetViewsRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     GetViewsRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: GetViewsRequest{
				Path:  "Test-Article-12-15",
				Year:  2023,
				Month: 12,
				Day:   15,
				Hour:  10,
			},
			wantErr: false,
		},
		{
			name:    "missing path",
			req:     GetViewsRequest{},
			wantErr: true,
			errMsg:  "path is required",
		},
		{
			name: "invalid year",
			req: GetViewsRequest{
				Path: "Test-Article-12-15",
				Year: 1999,
			},
			wantErr: true,
			errMsg:  "year must be between 2000 and 2100",
		},
		{
			name: "invalid month",
			req: GetViewsRequest{
				Path:  "Test-Article-12-15",
				Month: 13,
			},
			wantErr: true,
			errMsg:  "month must be between 1 and 12",
		},
		{
			name: "invalid day",
			req: GetViewsRequest{
				Path: "Test-Article-12-15",
				Day:  32,
			},
			wantErr: true,
			errMsg:  "day must be between 1 and 31",
		},
		{
			name: "invalid hour",
			req: GetViewsRequest{
				Path: "Test-Article-12-15",
				Hour: 25,
			},
			wantErr: true,
			errMsg:  "hour must be between 0 and 24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContentBuilder(t *testing.T) {
	t.Run("build simple content", func(t *testing.T) {
		content := NewContentBuilder().
			AddParagraph("Hello, World!").
			AddHeading("Section 1", 3).
			AddParagraph("This is a paragraph.").
			AddLink("Visit Example", "https://example.com").
			AddLineBreak().
			AddBlockquote("This is a quote.").
			AddCodeBlock("fmt.Println(\"Hello\")").
			Build()

		assert.Len(t, content, 7)
		
		// Check paragraph
		assert.Equal(t, "p", content[0].Tag)
		assert.Equal(t, "Hello, World!", content[0].Children[0].Content)
		
		// Check heading
		assert.Equal(t, "h3", content[1].Tag)
		assert.Equal(t, "Section 1", content[1].Children[0].Content)
		
		// Check link
		assert.Equal(t, "p", content[3].Tag)
		assert.Equal(t, "a", content[3].Children[0].Tag)
		assert.Equal(t, "https://example.com", content[3].Children[0].Attrs["href"])
		assert.Equal(t, "Visit Example", content[3].Children[0].Children[0].Content)
		
		// Check line break
		assert.Equal(t, "br", content[4].Tag)
		
		// Check blockquote
		assert.Equal(t, "blockquote", content[5].Tag)
		assert.Equal(t, "This is a quote.", content[5].Children[0].Content)
		
		// Check code block
		assert.Equal(t, "pre", content[6].Tag)
		assert.Equal(t, "fmt.Println(\"Hello\")", content[6].Children[0].Content)
	})
	
	t.Run("string representation", func(t *testing.T) {
		content := NewContentBuilder().
			AddParagraph("Hello").
			AddParagraph("World")
		
		str := content.String()
		assert.Contains(t, str, "Hello")
		assert.Contains(t, str, "World")
	})
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"", true},                                    // Empty string is valid
		{"https://example.com", true},                 // Valid HTTPS URL
		{"http://example.com", true},                  // Valid HTTP URL
		{"https://example.com/path", true},            // Valid URL with path
		{"https://example.com/path?query=value", true}, // Valid URL with query
		{"not-a-url", false},                          // Invalid URL
		{"ftp://example.com", false},                  // Invalid scheme
		{"https://", false},                           // Invalid URL
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			assert.Equal(t, tt.valid, isValidURL(tt.url))
		})
	}
}

func TestAPIError(t *testing.T) {
	t.Run("with code", func(t *testing.T) {
		err := &APIError{
			Code:        400,
			Description: "Bad Request",
		}
		assert.Equal(t, "Telegraph API error (code 400): Bad Request", err.Error())
	})
	
	t.Run("without code", func(t *testing.T) {
		err := &APIError{
			Description: "Something went wrong",
		}
		assert.Equal(t, "Telegraph API error: Something went wrong", err.Error())
	})
}