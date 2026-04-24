package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSMTPMailer_IsConfigured(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		port      int
		fromEmail string
		expected  bool
	}{
		{
			name:      "fully configured",
			host:      "smtp.gmail.com",
			port:      587,
			fromEmail: "noreply@example.com",
			expected:  true,
		},
		{
			name:      "missing host",
			host:      "",
			port:      587,
			fromEmail: "noreply@example.com",
			expected:  false,
		},
		{
			name:      "missing port",
			host:      "smtp.gmail.com",
			port:      0,
			fromEmail: "noreply@example.com",
			expected:  false,
		},
		{
			name:      "missing from email",
			host:      "smtp.gmail.com",
			port:      587,
			fromEmail: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := &SMTPMailer{
				host:      tt.host,
				port:      tt.port,
				fromEmail: tt.fromEmail,
			}
			assert.Equal(t, tt.expected, mailer.IsConfigured())
		})
	}
}

func TestSMTPMailer_FromName(t *testing.T) {
	tests := []struct {
		name     string
		fromName string
		expected string
	}{
		{
			name:     "custom from name",
			fromName: "Winx Team",
			expected: "Winx Team",
		},
		{
			name:     "default from name",
			fromName: "",
			expected: "Winx Notifications",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := &SMTPMailer{fromName: tt.fromName}
			assert.Equal(t, tt.expected, mailer.FromName())
		})
	}
}

func TestSMTPMailer_Send_Validation(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		port      int
		fromEmail string
		to        string
		wantErr   bool
	}{
		{
			name:      "valid configuration",
			host:      "smtp.test.com",
			port:      587,
			fromEmail: "noreply@test.com",
			to:        "user@test.com",
			wantErr:   false,
		},
		{
			name:      "missing host",
			host:      "",
			port:      587,
			fromEmail: "noreply@test.com",
			to:        "user@test.com",
			wantErr:   true,
		},
		{
			name:      "missing port",
			host:      "smtp.test.com",
			port:      0,
			fromEmail: "noreply@test.com",
			to:        "user@test.com",
			wantErr:   true,
		},
		{
			name:      "missing from email",
			host:      "smtp.test.com",
			port:      587,
			fromEmail: "",
			to:        "user@test.com",
			wantErr:   true,
		},
		{
			name:      "missing recipient",
			host:      "smtp.test.com",
			port:      587,
			fromEmail: "noreply@test.com",
			to:        "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := &SMTPMailer{
				host:      tt.host,
				port:      tt.port,
				fromEmail: tt.fromEmail,
			}
			err := mailer.Send(context.Background(), tt.to, "Test Subject", "Test Body")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
