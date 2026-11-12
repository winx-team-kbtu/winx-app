//go:build integration

package api

import (
	"context"
	"fmt"
	"os"
	"testing"

	"winx-notification/internal/app/domain/accountguard"
	userrepo "winx-notification/internal/app/domain/repositories/user"
	"winx-notification/internal/app/notifications"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type fakeMailer struct{}

func (fakeMailer) Send(context.Context, string, string, string) error { return nil }
func (fakeMailer) IsConfigured() bool                                 { return false }
func (fakeMailer) FromName() string                                   { return "Winx" }

// Requires WINX_TEST_DSN pointing at a database with all service migrations applied.
func consumerServer(t *testing.T) (*Server, *gorm.DB) {
	t.Helper()

	dsn := os.Getenv("WINX_TEST_DSN")
	if dsn == "" {
		t.Skip("WINX_TEST_DSN is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	cleanup := func() {
		db.Exec(`DELETE FROM notifications WHERE recipient LIKE 'cons%@test.io'`)
		db.Exec(`DELETE FROM users WHERE id BETWEEN 821 AND 824`)
	}
	cleanup()
	t.Cleanup(cleanup)

	for id := 821; id <= 823; id++ {
		require.NoError(t, db.Exec(`INSERT INTO users (id, email, password) VALUES (?, ?, 'x')`, id, fmt.Sprintf("cons%d@test.io", id)).Error)
	}

	return &Server{
		db:       db,
		mailer:   fakeMailer{},
		store:    notifications.NewStore(db),
		userRepo: userrepo.NewRepository(db),
		guard:    accountguard.New(db),
	}, db
}

func queued(t *testing.T, db *gorm.DB, email string) int64 {
	t.Helper()

	var n int64
	require.NoError(t, db.Raw(`SELECT count(*) FROM notifications WHERE recipient = ?`, email).Scan(&n).Error)

	return n
}

func registered(id int64) []byte {
	return []byte(fmt.Sprintf(`{"user_id":%d,"email":"cons%d@test.io","created_at":"2026-11-12T10:00:00Z"}`, id, id))
}

func TestUserRegisteredEvent_QueuesWelcomeForActiveAccount(t *testing.T) {
	s, db := consumerServer(t)

	require.NoError(t, s.handleUserRegistered(context.Background(), registered(821)))
	assert.EqualValues(t, 1, queued(t, db, "cons821@test.io"), "control: a live account still gets its welcome mail")
}

func TestUserRegisteredEvent_StaleEventForDeletedAccountStoresNothing(t *testing.T) {
	s, db := consumerServer(t)

	require.NoError(t, db.Exec(`DELETE FROM users WHERE id = 821`).Error)

	require.NoError(t, s.handleUserRegistered(context.Background(), registered(821)), "consumed, not retried forever")
	assert.Zero(t, queued(t, db, "cons821@test.io"), "no personal data may be recreated for an erased account")
}

func TestUserRegisteredEvent_StaleEventForPendingDeletionStoresNothing(t *testing.T) {
	s, db := consumerServer(t)

	require.NoError(t, db.Exec(`UPDATE users SET deletion_requested_at = NOW() WHERE id = 821`).Error)

	require.NoError(t, s.handleUserRegistered(context.Background(), registered(821)))
	assert.Zero(t, queued(t, db, "cons821@test.io"))
}

func TestUserRegisteredEvent_OldAccountEventDoesNotMailTheNewOwnerOfTheEmail(t *testing.T) {
	s, db := consumerServer(t)

	require.NoError(t, db.Exec(`DELETE FROM users WHERE id = 821`).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, email, password) VALUES (824, 'cons821@test.io', 'x')`).Error)

	require.NoError(t, s.handleUserRegistered(context.Background(), registered(821)))
	assert.Zero(t, queued(t, db, "cons821@test.io"), "the event belongs to account 821, which is gone")
}

func TestUserPasswordEvent_NoPinIsQueuedForDeletedOrPendingAccount(t *testing.T) {
	s, db := consumerServer(t)

	event := func(email string) []byte {
		return []byte(fmt.Sprintf(`{"email":%q,"pin_code":"123456","created_at":"2026-11-12T10:00:00Z"}`, email))
	}

	require.NoError(t, s.handleUserPassword(context.Background(), event("cons822@test.io")))
	assert.EqualValues(t, 1, queued(t, db, "cons822@test.io"), "control: a live account gets its PIN mail")

	require.NoError(t, db.Exec(`UPDATE users SET deletion_requested_at = NOW() WHERE id = 821`).Error)
	require.NoError(t, s.handleUserPassword(context.Background(), event("cons821@test.io")))
	assert.Zero(t, queued(t, db, "cons821@test.io"), "pending deletion")

	require.NoError(t, db.Exec(`DELETE FROM users WHERE id = 823`).Error)
	require.NoError(t, s.handleUserPassword(context.Background(), event("cons823@test.io")))
	assert.Zero(t, queued(t, db, "cons823@test.io"), "deleted")
}

func TestMatchCreatedEvent_OnlyActiveUsersAreNotified(t *testing.T) {
	s, db := consumerServer(t)

	require.NoError(t, db.Exec(`UPDATE users SET deletion_requested_at = NOW() WHERE id = 822`).Error)

	payload := []byte(`{"match_id":1,"user_one_id":821,"user_two_id":822,"created_at":"2026-11-12T10:00:00Z"}`)

	require.NoError(t, s.handleMatchCreated(context.Background(), payload))

	assert.EqualValues(t, 1, queued(t, db, "cons821@test.io"), "the active user is notified")
	assert.Zero(t, queued(t, db, "cons822@test.io"), "the user being deleted is not")
}
