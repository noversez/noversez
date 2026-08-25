package linguist

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const colorsURL = "https://raw.githubusercontent.com/github-linguist/linguist/refs/heads/main/lib/linguist/languages.yml"

var hexColor = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type Client struct {
	httpClient *http.Client
	url        string
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		url:        colorsURL,
	}
}

func (c *Client) Colors(ctx context.Context) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("User-Agent", "github-language-stats")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request Linguist colors: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("Linguist returned %s", resp.Status)
	}

	return ParseColors(resp.Body)
}

func ParseColors(reader io.Reader) (map[string]string, error) {
	colors := make(map[string]string)
	currentLanguage := ""

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if line[0] != ' ' && line[0] != '\t' {
			currentLanguage = ""
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") || trimmed == "---" || !strings.HasSuffix(trimmed, ":") {
				continue
			}
			currentLanguage = yamlScalar(strings.TrimSuffix(trimmed, ":"))
			continue
		}

		if currentLanguage == "" {
			continue
		}

		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "color:") {
			continue
		}

		color := yamlScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "color:")))
		if hexColor.MatchString(color) {
			colors[currentLanguage] = color
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read Linguist colors: %w", err)
	}
	if len(colors) == 0 {
		return nil, fmt.Errorf("Linguist color list is empty")
	}

	return colors, nil
}

func yamlScalar(value string) string {
	if len(value) >= 2 && value[0] == '"' {
		if unquoted, err := strconv.Unquote(value); err == nil {
			return unquoted
		}
	}
	if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
		return strings.ReplaceAll(value[1:len(value)-1], "''", "'")
	}
	return value
}
