package location

import (
	"context"

	"winx-profile/pkg/geoip"
)

type Service interface {
	Lookup(ctx context.Context, ip string) (geoip.Result, error)
}

type service struct {
	client geoip.Client
}

func NewService(client geoip.Client) Service {
	return &service{client: client}
}

func (s *service) Lookup(ctx context.Context, ip string) (geoip.Result, error) {
	return s.client.Lookup(ctx, ip)
}
