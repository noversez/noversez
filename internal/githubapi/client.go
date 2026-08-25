package githubapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL  = "https://api.github.com"
	defaultPageSize = 100
	defaultAttempts = 3
)

type Repository struct {
	FullName string `json:"full_name"`
	Fork     bool   `json:"fork"`
	Archived bool   `json:"archived"`
	Disabled bool   `json:"disabled"`
}

type Client struct {
	token       string
	baseURL     string
	httpClient  *http.Client
	pageSize    int
	maxAttempts int
	retryDelay  func(int) time.Duration
}

func New(token string) *Client {
	return &Client{
		token:       token,
		baseURL:     defaultBaseURL,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		pageSize:    defaultPageSize,
		maxAttempts: defaultAttempts,
		retryDelay: func(attempt int) time.Duration {
			return time.Duration(attempt) * time.Second
		},
	}
}

func (c *Client) ListRepositories(ctx context.Context) ([]Repository, error) {
	var all []Repository

	for page := 1; ; page++ {
		endpoint, err := url.Parse(c.baseURL + "/user/repos")
		if err != nil {
			return nil, fmt.Errorf("build repositories URL: %w", err)
		}

		query := endpoint.Query()
		query.Set("per_page", strconv.Itoa(c.pageSize))
		query.Set("page", strconv.Itoa(page))
		query.Set("affiliation", "owner,collaborator,organization_member")
		endpoint.RawQuery = query.Encode()

		var repos []Repository
		if err := c.getJSON(ctx, endpoint.String(), &repos); err != nil {
			return nil, fmt.Errorf("list repositories: %w", err)
		}

		all = append(all, repos...)
		if len(repos) < c.pageSize {
			return all, nil
		}
	}
}

func (c *Client) Languages(ctx context.Context, fullName string) (map[string]int64, error) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository name %q", fullName)
	}

	endpoint := fmt.Sprintf(
		"%s/repos/%s/%s/languages",
		c.baseURL,
		url.PathEscape(parts[0]),
		url.PathEscape(parts[1]),
	)

	languages := make(map[string]int64)
	if err := c.getJSON(ctx, endpoint, &languages); err != nil {
		return nil, err
	}
	return languages, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, target any) error {
	var lastErr error

	for attempt := 1; attempt <= c.maxAttempts; attempt++ {
		retry, err := c.request(ctx, endpoint, target)
		if err == nil {
			return nil
		}

		lastErr = err
		if !retry || attempt == c.maxAttempts {
			return lastErr
		}

		timer := time.NewTimer(c.retryDelay(attempt))
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return lastErr
}

func (c *Client) request(ctx context.Context, endpoint string, target any) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "github-language-stats")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return true, fmt.Errorf("request GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return false, fmt.Errorf("decode GitHub response: %w", err)
		}
		return false, nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	requestErr := fmt.Errorf(
		"GitHub returned %s: %s",
		resp.Status,
		strings.TrimSpace(string(body)),
	)
	retry := resp.StatusCode == http.StatusTooManyRequests ||
		resp.StatusCode >= http.StatusInternalServerError

	return retry, requestErr
}
