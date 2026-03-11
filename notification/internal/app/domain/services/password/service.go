package password

import (
	"winx-notification/configs"
	kafkacontract "winx-notification/internal/app/core/contracts/microservices/kafka-contract"
	repoDto "winx-notification/internal/app/domain/core/dto/repositories/password"
	eventdto "winx-notification/internal/app/domain/core/dto/services/event"
	dto "winx-notification/internal/app/domain/core/dto/services/password"
	"winx-notification/internal/app/domain/core/helpers/token"
	"winx-notification/internal/app/domain/core/helpers/validate"
	repository "winx-notification/internal/app/domain/repositories/password"
	tokenService "winx-notification/internal/app/domain/services/token"
	"winx-notification/internal/app/models/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrNotFound        = errors.New("token not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidToken    = errors.New("invalid token")
	ErrInvalidPinCode  = errors.New("invalid pincode")
	ErrTokenExpired    = errors.New("token expired")
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrInvalidUser     = errors.New("invalid user id")
	ErrInvalidPassword = errors.New("invalid password")
	ErrFailedPublish   = errors.New("failed to publish forgot password event")
)

type Service interface {
	ForgotPassword(ctx context.Context, input dto.ForgotPasswordDTO) error
	ResetPassword(ctx context.Context, input dto.ResetPasswordDTO) error
	ChangePassword(ctx context.Context, input dto.ChangePasswordDTO) error
	GetById(ctx context.Context, userID int64) (*models.User, error)
	VerifyPin(ctx context.Context, input dto.VerifyPinDTO) (models.PasswordReset, error)
}

type service struct {
	db           *gorm.DB
	repository   repository.Repository
	tokenService tokenService.Service
	kafka        kafkacontract.Producer
}

func NewService(
	db *gorm.DB,
	tokenService tokenService.Service,
	kafka kafkacontract.Producer,
) Service {
	return &service{
		db:           db,
		repository:   repository.NewRepository(db),
		tokenService: tokenService,
		kafka:        kafka,
	}
}

func (s *service) ForgotPassword(ctx context.Context, input dto.ForgotPasswordDTO) error {
	user, err := s.getByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUnauthenticated) {
			return ErrUnauthenticated
		}
		return err
	}

	pinCode := fmt.Sprintf("%06d", rand.Intn(1000000))
	randomToken := token.GenerateRandomToken(32)

	if err = s.repository.CreateResetToken(ctx, user.Email, pinCode, randomToken); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	payload, err := json.Marshal(eventdto.UserPasswordDTO{
		Email:     user.Email,
		PinCode:   pinCode,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return ErrFailedPublish
	}

	if err = s.kafka.Publish(ctx, configs.Config.Kafka.Topics.UserPassword, user.Email, payload); err != nil {
		return ErrFailedPublish
	}

	return nil
}

func (s *service) ResetPassword(ctx context.Context, input dto.ResetPasswordDTO) error {
	user, err := s.repository.GetUserByResetToken(ctx, repoDto.ResetPasswordDTO{
		Email: input.Email,
		Token: input.Token,
	})
	if err != nil {
		return err
	}

	if input.NewPasswordConfirmation != input.NewPassword {
		return ErrInvalidPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	if err := s.repository.UpdatePassword(ctx, user.ID, string(hashed)); err != nil {
		return err
	}

	if err := s.repository.DeleteResetToken(ctx, input.Token); err != nil {
		return err
	}

	return err
}

func (s *service) ChangePassword(ctx context.Context, input dto.ChangePasswordDTO) error {
	tokenInfo, err := s.tokenService.ValidateToken(ctx, input.Token)
	if err != nil {
		return ErrInvalidUser
	}

	u64, err := strconv.ParseUint(tokenInfo.GetUserID(), 10, 64)
	if err != nil {
		return ErrInvalidUser
	}

	userID := int64(u64)
	user, err := s.GetById(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	err = validate.Validate(input, user)
	if err != nil {
		return ErrInvalidPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	err = s.repository.UpdatePassword(ctx, user.ID, string(hashedPassword))
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetById(ctx context.Context, userID int64) (*models.User, error) {
	user, err := s.repository.GetById(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUnauthenticated
		}
	}
	return user, nil
}

func (s *service) VerifyPin(ctx context.Context, input dto.VerifyPinDTO) (models.PasswordReset, error) {
	reset, err := s.repository.GetUserByPinCode(ctx, repoDto.VerifyPinDTO{
		Email:   input.Email,
		PinCode: input.PinCode,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return reset, ErrUnauthenticated
		}
		if errors.Is(err, repository.ErrInvalidPinCode) {
			return reset, ErrInvalidToken
		}
	}

	return reset, nil
}

func (s *service) getByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.repository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUnauthenticated) {
			return nil, ErrUnauthenticated
		}
		return nil, err
	}
	return user, nil
}
