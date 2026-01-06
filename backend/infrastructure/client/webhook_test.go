package client

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
)

// TestAlarm tests the Alarm method with various scenarios using gomonkey
func TestAlarm(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		message      string
		mockResponse *http.Response
		mockError    error
		wantErr      bool
		errContains  string
		setupNewReq  bool
		newReqError  error
	}{
		{
			name:    "successful alarm with custom message",
			url:     "https://ntfy.sh/test-topic",
			message: "Test alarm message",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("ok")),
			},
			mockError: nil,
			wantErr:   false,
		},
		{
			name:    "successful alarm with empty message (default)",
			url:     "https://ntfy.sh/test-topic",
			message: "",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("ok")),
			},
			mockError: nil,
			wantErr:   false,
		},
		{
			name:        "empty URL returns error",
			url:         "",
			message:     "Test message",
			wantErr:     true,
			errContains: "ntfy URL is empty",
		},
		{
			name:    "server returns 4xx error",
			url:     "https://ntfy.sh/test-topic",
			message: "Test message",
			mockResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString("bad request")),
			},
			mockError:   nil,
			wantErr:     true,
			errContains: "ntfy returned non-2xx status: 400",
		},
		{
			name:    "server returns 5xx error",
			url:     "https://ntfy.sh/test-topic",
			message: "Test message",
			mockResponse: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString("internal error")),
			},
			mockError:   nil,
			wantErr:     true,
			errContains: "ntfy returned non-2xx status: 500",
		},
		{
			name:    "successful alarm with 201 status",
			url:     "https://ntfy.sh/test-topic",
			message: "Test message",
			mockResponse: &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(bytes.NewBufferString("created")),
			},
			mockError: nil,
			wantErr:   false,
		},
		{
			name:        "http request fails",
			url:         "https://ntfy.sh/test-topic",
			message:     "Test message",
			mockError:   errors.New("connection refused"),
			wantErr:     true,
			errContains: "request failed",
		},
		{
			name:        "invalid URL causes NewRequest to fail",
			url:         "ht!tp://invalid url",
			message:     "Test message",
			setupNewReq: true,
			newReqError: errors.New("invalid URL escape"),
			wantErr:     true,
			errContains: "failed to create request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewWebhookClient()

			// Skip mocking for empty URL test
			if tt.url == "" {
				err := client.Alarm(tt.url, tt.message)
				if (err != nil) != tt.wantErr {
					t.Errorf("Alarm() error = %v, wantErr %v", err, tt.wantErr)
				}
				if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Alarm() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			// Mock http.NewRequest if needed
			if tt.setupNewReq {
				patches := gomonkey.ApplyFunc(http.NewRequest, func(method, url string, body io.Reader) (*http.Request, error) {
					return nil, tt.newReqError
				})
				defer patches.Reset()

				err := client.Alarm(tt.url, tt.message)
				if (err != nil) != tt.wantErr {
					t.Errorf("Alarm() error = %v, wantErr %v", err, tt.wantErr)
				}
				if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Alarm() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			// Mock http.Client.Do method
			patches := gomonkey.ApplyMethod(&http.Client{}, "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
				// Verify request method
				if req.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", req.Method)
				}

				// Verify URL
				if req.URL.String() != tt.url {
					t.Errorf("Expected URL %s, got %s", tt.url, req.URL.String())
				}

				// Verify headers (only if we have a successful mock response)
				if tt.mockResponse != nil && tt.mockError == nil {
					if req.Header.Get("Title") != "Offline Me Alarm" {
						t.Errorf("Expected Title header 'Offline Me Alarm', got '%s'", req.Header.Get("Title"))
					}
					if req.Header.Get("Priority") != "urgent" {
						t.Errorf("Expected Priority header 'urgent', got '%s'", req.Header.Get("Priority"))
					}
					if req.Header.Get("Tags") != "warning,skull" {
						t.Errorf("Expected Tags header 'warning,skull', got '%s'", req.Header.Get("Tags"))
					}
					if req.Header.Get("Content-Type") != "text/plain" {
						t.Errorf("Expected Content-Type 'text/plain', got '%s'", req.Header.Get("Content-Type"))
					}

					// Verify message body
					if req.Body != nil {
						bodyBytes, _ := io.ReadAll(req.Body)
						expectedMessage := tt.message
						if expectedMessage == "" {
							expectedMessage = "Work time notification"
						}
						if string(bodyBytes) != expectedMessage {
							t.Errorf("Expected body '%s', got '%s'", expectedMessage, string(bodyBytes))
						}
					}
				}

				return tt.mockResponse, tt.mockError
			})
			defer patches.Reset()

			// Call the Alarm method
			err := client.Alarm(tt.url, tt.message)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Alarm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check error message content
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Alarm() error = %v, want error containing %v", err, tt.errContains)
			}
		})
	}
}
