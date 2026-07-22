package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coollabsio/coolify-cli/internal/api"
)

func TestMCPService_Enable(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		wantErr    bool
		wantMsg    string
	}{
		{
			name:       "successful enable",
			response:   `{"message": "MCP server enabled."}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantMsg:    "MCP server enabled.",
		},
		{
			name:       "forbidden",
			response:   `{"message": "This action is unauthorized."}`,
			statusCode: http.StatusForbidden,
			wantErr:    true,
		},
		{
			name:       "server error",
			response:   `{"error": "internal server error"}`,
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected method POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/mcp/enable" {
					t.Errorf("Expected path /api/v1/mcp/enable, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			svc := NewMCPService(client)

			resp, err := svc.Enable(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("MCPService.Enable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp.Message != tt.wantMsg {
				t.Errorf("MCPService.Enable() message = %q, want %q", resp.Message, tt.wantMsg)
			}
		})
	}
}

func TestMCPService_Disable(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		wantErr    bool
		wantMsg    string
	}{
		{
			name:       "successful disable",
			response:   `{"message": "MCP server disabled."}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantMsg:    "MCP server disabled.",
		},
		{
			name:       "forbidden",
			response:   `{"message": "This action is unauthorized."}`,
			statusCode: http.StatusForbidden,
			wantErr:    true,
		},
		{
			name:       "server error",
			response:   `{"error": "internal server error"}`,
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected method POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/mcp/disable" {
					t.Errorf("Expected path /api/v1/mcp/disable, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			svc := NewMCPService(client)

			resp, err := svc.Disable(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("MCPService.Disable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp.Message != tt.wantMsg {
				t.Errorf("MCPService.Disable() message = %q, want %q", resp.Message, tt.wantMsg)
			}
		})
	}
}
