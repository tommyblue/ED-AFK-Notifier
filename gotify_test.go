package notifier

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tommyblue/ED-AFK-Notifier/bots"
)

func TestGotify_New(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		token       string
		title       string
		priority    int
		expectError bool
	}{
		{
			name:        "Valid configuration",
			url:         "https://gotify.example.com",
			token:       "abc123",
			title:       "Test Title",
			priority:    5,
			expectError: false,
		},
		{
			name:        "Empty URL",
			url:         "",
			token:       "abc123",
			title:       "Test Title",
			priority:    5,
			expectError: true,
		},
		{
			name:        "Empty token",
			url:         "https://gotify.example.com",
			token:       "",
			title:       "Test Title",
			priority:    5,
			expectError: true,
		},
		{
			name:        "Negative priority",
			url:         "https://gotify.example.com",
			token:       "abc123",
			title:       "Test Title",
			priority:    -1,
			expectError: false, // Should allow negative priority
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bots.NewGotify(tt.url, tt.token, tt.title, tt.priority)
			if (err != nil) != tt.expectError {
				t.Errorf("NewGotify() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestGotify_Send(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		serverResponse int
		expectError    bool
	}{
		{
			name:           "Successful notification",
			message:        "Test message",
			serverResponse: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Server error",
			message:        "Test message",
			serverResponse: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "Unauthorized error",
			message:        "Test message",
			serverResponse: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name:           "Empty message",
			message:        "",
			serverResponse: http.StatusOK,
			expectError:    false, // Still should send, just empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server that responds with the specified status code
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request headers and body
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				// Return the configured status code
				w.WriteHeader(tt.serverResponse)
			}))
			defer server.Close()

			// Create a Gotify instance with our test server
			g, err := bots.NewGotify(server.URL, "test-token", "Test Title", 5)
			if err != nil {
				t.Fatalf("Failed to create Gotify instance: %v", err)
			}

			// Send the message
			err = g.Send(tt.message)

			// Check if error matches expectations
			if (err != nil) != tt.expectError {
				t.Errorf("Send() error = %v, expectError %v", err, tt.expectError)
			}

			// Verify error message if expected
			if tt.expectError && tt.serverResponse != http.StatusOK && err != nil {
				expected := fmt.Sprintf("unexpected status code: %d", tt.serverResponse)
				if err.Error() != expected {
					t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
				}
			}
		})
	}
}
