package match

import (
	"context"
	"errors"
	"fmt"
	repodto "winx-match/internal/app/domain/core/dto/repositories/match"
	matchServicedto "winx-match/internal/app/domain/core/dto/services/match"
	matchRepo "winx-match/internal/app/domain/repositories/match"
	"winx-match/internal/app/models/models"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("match not found")

// Service — интерфейс сервиса матчей.
// Добавляй новые методы сюда и реализуй в service struct.
type Service interface {
	List(ctx context.Context, input matchServicedto.ListDTO) ([]matchServicedto.MatchWithProfile, error)
	Create(ctx context.Context, user1ID, user2ID int64) (models.Match, error)
	Delete(ctx context.Context, input matchServicedto.DeleteDTO) (bool, error)
}

type service struct {
	matchRepo matchRepo.Repository
}

func NewService(db *gorm.DB) Service {
	return &service{
		matchRepo: matchRepo.NewRepository(db),
	}
}

func (s *service) List(ctx context.Context, input matchServicedto.ListDTO) ([]matchServicedto.MatchWithProfile, error) {
	limit := input.Limit
	if limit == 0 {
		limit = 20
	}

	rows, err := s.matchRepo.ListWithProfiles(ctx, repodto.ListDTO{
		UserID: input.UserID,
		Limit:  limit,
		Offset: input.Offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]matchServicedto.MatchWithProfile, 0, len(rows))
	for _, r := range rows {
		result = append(result, matchServicedto.MatchWithProfile{
			ID:            r.ID,
			MatchedUserID: r.MatchedUserID,
			Name:          r.Name,
			City:          r.City,
			Country:       r.Country,
			PhotoURL:      r.PhotoURL,
			CreatedAt:     r.CreatedAt,
		})
	}
	return result, nil
}

func (s *service) Delete(ctx context.Context, input matchServicedto.DeleteDTO) (bool, error) {
	ok, err := s.matchRepo.DeleteByIDAndUserID(ctx, input.ID, input.UserID)
	if err != nil {
		if errors.Is(err, matchRepo.ErrNotFound) {
			return false, ErrNotFound
		}
		return false, err
	}
	return ok, nil
}

func (s *service) Create(ctx context.Context, user1ID, user2ID int64) (models.Match, error) {
	match, err := s.matchRepo.Create(ctx, repodto.CreateDTO{
		UserOneID: user1ID,
		UserTwoID: user2ID,
	})
	if err != nil {
		return models.Match{}, fmt.Errorf("create match: %w", err)
	}
	return match, nil
}
