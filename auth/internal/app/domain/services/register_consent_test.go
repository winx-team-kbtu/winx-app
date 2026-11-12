package services

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"auth/configs"
	dto "auth/internal/app/domain/core/dto/services"
	userDto "auth/internal/app/domain/core/dto/services/user"
	"auth/internal/app/models/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type consentUserService struct {
	created []userDto.CreateDTO
	err     error
}

func (f *consentUserService) Create(_ context.Context, in userDto.CreateDTO) (models.User, error) {
	f.created = append(f.created, in)
	if f.err != nil {
		return models.User{}, f.err
	}

	return models.User{ID: 1, Email: in.Email}, nil
}

func (f *consentUserService) Delete(context.Context, userDto.DeleteDTO) (bool, error) {
	return false, nil
}

func (f *consentUserService) Update(context.Context, userDto.UpdateDTO) (models.User, error) {
	return models.User{}, nil
}

func (f *consentUserService) GetByEmail(context.Context, string) (models.User, error) {
	return models.User{}, nil
}

type consentKafka struct{ published int }

func (k *consentKafka) Publish(context.Context, string, string, []byte) error {
	k.published++

	return nil
}

func (k *consentKafka) Close() error { return nil }

// newConsentService builds a service with the given fakes. The config global is
// normally loaded at startup, so tests give it an empty value.
func newConsentService(users *consentUserService, kafka *consentKafka) *service {
	if configs.Config == nil {
		// The config type is unexported, so allocate a zero value through reflection.
		v := reflect.ValueOf(&configs.Config).Elem()
		v.Set(reflect.New(v.Type().Elem()))
	}

	return &service{userService: users, kafka: kafka}
}

func TestRegister_RequiresConsent(t *testing.T) {
	users := &consentUserService{}
	kafka := &consentKafka{}
	svc := newConsentService(users, kafka)

	_, err := svc.Register(context.Background(), dto.RegisterDTO{Email: "a@x.io", Password: "pw", ConsentAccepted: false})

	assert.ErrorIs(t, err, ErrConsentRequired)
	assert.Empty(t, users.created, "no account may be created without consent")
	assert.Zero(t, kafka.published, "no registration event without consent")
}

func TestRegister_PersistsNoticeVersionAndTimestamp(t *testing.T) {
	users := &consentUserService{}
	svc := newConsentService(users, &consentKafka{})

	before := time.Now().UTC().Add(-time.Second)

	_, err := svc.Register(context.Background(), dto.RegisterDTO{Email: "a@x.io", Password: "pw", ConsentAccepted: true})
	require.NoError(t, err)
	require.Len(t, users.created, 1)

	got := users.created[0]
	assert.Equal(t, ConsentNoticeVersion, got.ConsentNoticeVersion)
	assert.True(t, got.ConsentAcceptedAt.After(before), "timestamp must be the acceptance time")
	assert.False(t, got.ConsentAcceptedAt.After(time.Now().UTC().Add(time.Second)))
	assert.Equal(t, time.UTC, got.ConsentAcceptedAt.Location())
}

func TestRegister_PublishesEventOnlyAfterConsentedRegistration(t *testing.T) {
	kafka := &consentKafka{}
	svc := newConsentService(&consentUserService{}, kafka)

	_, err := svc.Register(context.Background(), dto.RegisterDTO{Email: "a@x.io", Password: "pw", ConsentAccepted: true})
	require.NoError(t, err)
	assert.Equal(t, 1, kafka.published)
}

func TestRegister_UserCreationErrorIsPropagated(t *testing.T) {
	boom := errors.New("db down")
	svc := newConsentService(&consentUserService{err: boom}, &consentKafka{})

	_, err := svc.Register(context.Background(), dto.RegisterDTO{Email: "a@x.io", Password: "pw", ConsentAccepted: true})
	assert.ErrorIs(t, err, boom)
}
