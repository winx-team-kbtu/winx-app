package profileclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client calls the internal API of the profile service.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func New(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http:    &http.Client{Timeout: timeout},
	}
}

// DeleteUserData asks the profile service to remove everything it holds for the user.
// The call is idempotent on the profile side, so it is safe to repeat. Only a 2xx answer
// counts as success.
func (c *Client) DeleteUserData(ctx context.Context, userID int64) error {
	url := fmt.Sprintf("%s/internal/users/%d", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("build profile cleanup request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call profile cleanup: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))

		return fmt.Errorf("profile cleanup returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}
