package githubapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestClientListsPagesAndLoadsLanguages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected Authorization header")
		}

		switch r.URL.Path {
		case "/user/repos":
			if got := r.URL.Query().Get("affiliation"); got != "owner,collaborator,organization_member" {
				t.Errorf("affiliation = %q", got)
			}

			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			if page == 1 {
				_ = json.NewEncoder(w).Encode([]Repository{
					{FullName: "me/one"},
					{FullName: "me/two", Fork: true},
				})
				return
			}
			_ = json.NewEncoder(w).Encode([]Repository{{FullName: "me/three"}})
		case "/repos/me/one/languages":
			_ = json.NewEncoder(w).Encode(map[string]int64{"Go": 42})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New("test-token")
	client.baseURL = server.URL
	client.httpClient = server.Client()
	client.pageSize = 2

	repos, err := client.ListRepositories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 3 {
		t.Fatalf("got %d repositories, want 3", len(repos))
	}

	languages, err := client.Languages(context.Background(), "me/one")
	if err != nil {
		t.Fatal(err)
	}
	if languages["Go"] != 42 {
		t.Fatalf("Go bytes = %d, want 42", languages["Go"])
	}
}

func TestClientRetriesServerErrors(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts == 1 {
			http.Error(w, "temporary", http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode([]Repository{})
	}))
	defer server.Close()

	client := New("test-token")
	client.baseURL = server.URL
	client.httpClient = server.Client()
	client.retryDelay = func(int) time.Duration { return 0 }

	if _, err := client.ListRepositories(context.Background()); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}
