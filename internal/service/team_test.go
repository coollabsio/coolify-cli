package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coollabsio/coolify-cli/internal/api"
)

func TestTeamService_List(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name: "successful list",
			response: `[
				{"uuid": "team-1", "name": "Team 1", "description": "First team"},
				{"uuid": "team-2", "name": "Team 2", "description": "Second team"}
			]`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:       "empty list",
			response:   `[]`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  0,
		},
		{
			name:       "server error",
			response:   `{"error": "internal server error"}`,
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/teams" {
					t.Errorf("Expected path /api/v1/teams, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			svc := NewTeamService(client)

			teams, err := svc.List(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("TeamService.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(teams) != tt.wantCount {
				t.Errorf("TeamService.List() got %d teams, want %d", len(teams), tt.wantCount)
			}
		})
	}
}

func TestTeamService_Get(t *testing.T) {
	tests := []struct {
		name       string
		teamID     string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful get",
			teamID:     "1",
			response:   `{"id": 1, "name": "Test Team", "description": "A test team", "personal_team": false, "created_at": "2025-01-01", "updated_at": "2025-01-01", "show_boarding": false}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "not found",
			teamID:     "nonexistent",
			response:   `{"error": "team not found"}`,
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/api/v1/teams/" + tt.teamID
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			svc := NewTeamService(client)

			team, err := svc.Get(context.Background(), tt.teamID)

			if (err != nil) != tt.wantErr {
				t.Errorf("TeamService.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && team.ID == 0 {
				t.Errorf("TeamService.Get() got empty team ID")
			}
		})
	}
}

func TestTeamService_Current(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		wantErr    bool
		wantName   string
	}{
		{
			name:       "successful get current",
			response:   `{"uuid": "current-team", "name": "Current Team", "description": "The current team"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantName:   "Current Team",
		},
		{
			name:       "unauthorized",
			response:   `{"error": "unauthorized"}`,
			statusCode: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/teams/current" {
					t.Errorf("Expected path /api/v1/teams/current, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			svc := NewTeamService(client)

			team, err := svc.Current(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("TeamService.Current() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && team.Name != tt.wantName {
				t.Errorf("TeamService.Current() got name %s, want %s", team.Name, tt.wantName)
			}
		})
	}
}

func TestTeamService_ListMembers(t *testing.T) {
	tests := []struct {
		name       string
		teamID     string
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name:   "successful list members",
			teamID: "team-123",
			response: `[
				{"uuid": "user-1", "name": "Alice", "email": "alice@example.com", "role": "admin"},
				{"uuid": "user-2", "name": "Bob", "email": "bob@example.com", "role": "member"}
			]`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:       "empty members list",
			teamID:     "team-456",
			response:   `[]`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/api/v1/teams/" + tt.teamID + "/members"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			svc := NewTeamService(client)

			members, err := svc.ListMembers(context.Background(), tt.teamID)

			if (err != nil) != tt.wantErr {
				t.Errorf("TeamService.ListMembers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(members) != tt.wantCount {
				t.Errorf("TeamService.ListMembers() got %d members, want %d", len(members), tt.wantCount)
			}
		})
	}
}

func TestTeamService_CurrentMembers(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name: "successful list current members",
			response: `[
				{"uuid": "user-1", "name": "Alice", "email": "alice@example.com"},
				{"uuid": "user-2", "name": "Bob", "email": "bob@example.com"}
			]`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/teams/current/members" {
					t.Errorf("Expected path /api/v1/teams/current/members, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			svc := NewTeamService(client)

			members, err := svc.CurrentMembers(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("TeamService.CurrentMembers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(members) != tt.wantCount {
				t.Errorf("TeamService.CurrentMembers() got %d members, want %d", len(members), tt.wantCount)
			}

			// Verify JSON marshaling works
			if !tt.wantErr && len(members) > 0 {
				_, err := json.Marshal(members[0])
				if err != nil {
					t.Errorf("Failed to marshal team member: %v", err)
				}
			}
		})
	}
}
