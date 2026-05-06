package email

import "context"

// Mailer is the interface both SMTPMailer and ResendMailer implement.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
	IsConfigured() bool
	FromName() string
}
