package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const resendAPIURL = "https://api.resend.com/emails"

type ResendMailer struct {
	apiKey    string
	fromEmail string
	fromName  string
}

func NewResendMailer(apiKey, fromEmail, fromName string) *ResendMailer {
	return &ResendMailer{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

func (m *ResendMailer) IsConfigured() bool {
	return strings.TrimSpace(m.apiKey) != ""
}

func (m *ResendMailer) FromName() string {
	if strings.TrimSpace(m.fromName) != "" {
		return m.fromName
	}
	return "Winx Notifications"
}

func (m *ResendMailer) Send(ctx context.Context, to, subject, body string) error {
	payload := map[string]any{
		"from":    fmt.Sprintf("%s <%s>", m.FromName(), m.fromEmail),
		"to":      []string{to},
		"subject": subject,
		"text":    body,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal resend payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIURL, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("build resend request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("resend send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API %d: %s", resp.StatusCode, string(raw))
	}

	return nil
}
