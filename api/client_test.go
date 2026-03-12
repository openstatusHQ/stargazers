package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/shurcooL/githubv4"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	ghClient := githubv4.NewEnterpriseClient(server.URL, server.Client())
	return &Client{client: ghClient}
}

func graphqlResponse(data string) string {
	return `{"data":` + data + `}`
}

func stargazersResponse(users []map[string]any, hasNextPage bool, endCursor string, totalCount int) string {
	nodes, _ := json.Marshal(users)
	data := map[string]any{
		"repository": map[string]any{
			"name": "testrepo",
			"stargazers": map[string]any{
				"totalCount": totalCount,
				"pageInfo": map[string]any{
					"endCursor":   endCursor,
					"hasNextPage": hasNextPage,
				},
				"nodes": json.RawMessage(nodes),
			},
		},
	}
	b, _ := json.Marshal(data)
	return graphqlResponse(string(b))
}

func TestGetStargazers_SinglePage(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		users := []map[string]any{
			{
				"avatarUrl": "https://avatar1.url",
				"bio":       "dev1",
				"company":   "@acme",
				"email":     "alice@example.com",
				"followers": map[string]any{"totalCount": 100},
				"following": map[string]any{"totalCount": 50},
				"socialAccounts": map[string]any{
					"nodes":      []any{},
					"totalCount": 0,
				},
				"login": "alice",
				"name":  "Alice",
			},
			{
				"avatarUrl": "https://avatar2.url",
				"bio":       "dev2",
				"company":   "",
				"email":     "bob@example.com",
				"followers": map[string]any{"totalCount": 200},
				"following": map[string]any{"totalCount": 75},
				"socialAccounts": map[string]any{
					"nodes":      []any{},
					"totalCount": 0,
				},
				"login": "bob",
				"name":  "Bob",
			},
		}
		resp := stargazersResponse(users, false, "cursor_end", 2)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(resp))
	})

	users, cursor, err := client.GetStargazers("org", "repo", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Login != "alice" {
		t.Errorf("expected login 'alice', got %q", users[0].Login)
	}
	if users[0].Followers != 100 {
		t.Errorf("expected 100 followers, got %d", users[0].Followers)
	}
	if users[1].Login != "bob" {
		t.Errorf("expected login 'bob', got %q", users[1].Login)
	}
	if cursor != "cursor_end" {
		t.Errorf("expected cursor 'cursor_end', got %q", cursor)
	}
}

func TestGetStargazers_MultiPage(t *testing.T) {
	var callCount atomic.Int32

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")

		if n == 1 {
			users := []map[string]any{
				{
					"avatarUrl": "", "bio": "", "company": "", "email": "",
					"followers":      map[string]any{"totalCount": 0},
					"following":      map[string]any{"totalCount": 0},
					"socialAccounts": map[string]any{"nodes": []any{}, "totalCount": 0},
					"login":          "alice",
					"name":           "Alice",
				},
			}
			w.Write([]byte(stargazersResponse(users, true, "page2_cursor", 2)))
		} else {
			users := []map[string]any{
				{
					"avatarUrl": "", "bio": "", "company": "", "email": "",
					"followers":      map[string]any{"totalCount": 0},
					"following":      map[string]any{"totalCount": 0},
					"socialAccounts": map[string]any{"nodes": []any{}, "totalCount": 0},
					"login":          "bob",
					"name":           "Bob",
				},
			}
			w.Write([]byte(stargazersResponse(users, false, "final_cursor", 2)))
		}
	})

	users, cursor, err := client.GetStargazers("org", "repo", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users across pages, got %d", len(users))
	}
	if users[0].Login != "alice" || users[1].Login != "bob" {
		t.Errorf("unexpected users: %+v", users)
	}
	if cursor != "final_cursor" {
		t.Errorf("expected cursor 'final_cursor', got %q", cursor)
	}
	if callCount.Load() != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount.Load())
	}
}

func TestGetStargazers_WithStartCursor(t *testing.T) {
	var receivedVars map[string]any

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Variables map[string]any `json:"variables"`
		}
		json.Unmarshal(body, &req)
		receivedVars = req.Variables

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(stargazersResponse(nil, false, "", 0)))
	})

	client.GetStargazers("org", "repo", "my_cursor_123")

	if receivedVars["after"] != "my_cursor_123" {
		t.Errorf("expected after='my_cursor_123', got %v", receivedVars["after"])
	}
}

func TestGetStargazers_APIError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	})

	_, _, err := client.GetStargazers("org", "repo", "")
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestGetStargazers_SocialAccounts(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		users := []map[string]any{
			{
				"avatarUrl": "", "bio": "", "company": "", "email": "",
				"followers": map[string]any{"totalCount": 0},
				"following": map[string]any{"totalCount": 0},
				"socialAccounts": map[string]any{
					"nodes": []any{
						map[string]any{"provider": "LINKEDIN", "url": "https://linkedin.com/in/alice"},
						map[string]any{"provider": "BLUESKY", "url": "https://bsky.app/profile/alice"},
					},
					"totalCount": 2,
				},
				"login": "alice",
				"name":  "Alice",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(stargazersResponse(users, false, "end", 1)))
	})

	users, _, err := client.GetStargazers("org", "repo", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].LinkedinUrl != "https://linkedin.com/in/alice" {
		t.Errorf("expected LinkedIn URL, got %q", users[0].LinkedinUrl)
	}
	if users[0].BskyUrl != "https://bsky.app/profile/alice" {
		t.Errorf("expected Bluesky URL, got %q", users[0].BskyUrl)
	}
}

func TestGetCompany_Success(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"organization": map[string]any{
				"name":        "Acme Inc",
				"login":       "acme",
				"avatarUrl":   "https://avatar.acme.com",
				"description": "Building things",
				"email":       "info@acme.com",
				"location":    "San Francisco",
				"membersWithRole": map[string]any{
					"totalCount": 42,
				},
				"repositories": map[string]any{
					"totalCount": 100,
				},
				"websiteUrl": "https://acme.com",
			},
		}
		b, _ := json.Marshal(data)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(graphqlResponse(string(b))))
	})

	company, err := client.GetCompany("acme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if company.Name != "Acme Inc" {
		t.Errorf("expected name 'Acme Inc', got %q", company.Name)
	}
	if company.Login != "acme" {
		t.Errorf("expected login 'acme', got %q", company.Login)
	}
	if company.Members != 42 {
		t.Errorf("expected 42 members, got %d", company.Members)
	}
	if company.Repositories != 100 {
		t.Errorf("expected 100 repos, got %d", company.Repositories)
	}
	if company.WebsiteUrl != "https://acme.com" {
		t.Errorf("expected website 'https://acme.com', got %q", company.WebsiteUrl)
	}
}

func TestGetCompany_NotFound(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{},"errors":[{"message":"Could not resolve to an Organization","type":"NOT_FOUND","path":["organization"]}]}`))
	})

	_, err := client.GetCompany("nonexistent")
	if err == nil {
		t.Fatal("expected error for not found org, got nil")
	}
}
