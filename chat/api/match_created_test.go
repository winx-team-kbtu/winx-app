package api

import (
	"context"
	"errors"
	"testing"

	servicedto "winx-chat/internal/app/domain/core/dto/services/chat"
	"winx-chat/internal/app/models/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeChatService struct {
	created [][3]int64
}

func (f *fakeChatService) CreateFromMatch(_ context.Context, matchID, one, two int64) (models.Chat, error) {
	f.created = append(f.created, [3]int64{matchID, one, two})

	return models.Chat{ID: 1, MatchID: matchID}, nil
}

func (f *fakeChatService) List(context.Context, servicedto.ListDTO) ([]models.Chat, error) {
	return nil, nil
}

func (f *fakeChatService) FindByID(context.Context, int64) (models.Chat, error) {
	return models.Chat{}, nil
}

type fakeGuard struct {
	active bool
	err    error
	asked  [][]int64
}

func (g *fakeGuard) UsersActive(_ context.Context, ids ...int64) (bool, error) {
	g.asked = append(g.asked, ids)

	return g.active, g.err
}

const matchPayload = `{"match_id":7,"user_one_id":11,"user_two_id":22,"created_at":"2026-11-12T10:00:00Z"}`

func TestHandleMatchCreated_CreatesChatForActiveUsers(t *testing.T) {
	chats, guard := &fakeChatService{}, &fakeGuard{active: true}
	s := &Server{chatService: chats, guard: guard}

	require.NoError(t, s.handleMatchCreated(context.Background(), []byte(matchPayload)))

	assert.Equal(t, [][3]int64{{7, 11, 22}}, chats.created)
	assert.Equal(t, [][]int64{{11, 22}}, guard.asked, "both users of the match are checked")
}

func TestHandleMatchCreated_StaleEventForDeletedUserCreatesNothing(t *testing.T) {
	chats, guard := &fakeChatService{}, &fakeGuard{active: false}
	s := &Server{chatService: chats, guard: guard}

	err := s.handleMatchCreated(context.Background(), []byte(matchPayload))

	require.NoError(t, err, "a stale event is consumed (committed), not retried forever")
	assert.Empty(t, chats.created, "no chat may be recreated for an erased account")
}

func TestHandleMatchCreated_GuardFailureIsNotTreatedAsActive(t *testing.T) {
	chats, guard := &fakeChatService{}, &fakeGuard{active: true, err: errors.New("db down")}
	s := &Server{chatService: chats, guard: guard}

	err := s.handleMatchCreated(context.Background(), []byte(matchPayload))

	require.Error(t, err, "if account state is unknown the event must not be committed")
	assert.Empty(t, chats.created, "and nothing is created on a guess")
}

func TestHandleMatchCreated_InvalidPayloadIsAnError(t *testing.T) {
	chats, guard := &fakeChatService{}, &fakeGuard{active: true}
	s := &Server{chatService: chats, guard: guard}

	require.Error(t, s.handleMatchCreated(context.Background(), []byte(`{not json`)))
	assert.Empty(t, guard.asked)
	assert.Empty(t, chats.created)
}
