package geoip

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"winx-profile/configs"
	corehttp "winx-profile/internal/app/core/http"
)

var ErrLookupFailed = errors.New("failed to lookup geo info by ip")

type Result struct {
	IP        string
	City      string
	Country   string
	Latitude  float64
	Longitude float64
}

type Client interface {
	Lookup(ctx context.Context, ip string) (Result, error)
}

type client struct {
	baseURL string
	http    *corehttp.ClientBase
}

type response struct {
	IP          string  `json:"ip"`
	City        string  `json:"city"`
	CountryName string  `json:"country_name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Error       bool    `json:"error"`
	Reason      string  `json:"reason"`
	Message     string  `json:"message"`
}

func NewClient() Client {
	timeout := time.Duration(configs.Config.GeoIP.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &client{
		baseURL: strings.TrimRight(configs.Config.GeoIP.BaseURL, "/"),
		http:    corehttp.New(corehttp.ClientConfig{Timeout: timeout}),
	}
}

func (c *client) Lookup(ctx context.Context, ip string) (Result, error) {
	ip = strings.TrimSpace(ip)

	url := fmt.Sprintf("%s/json/", c.baseURL)
	if ip != "" {
		url = fmt.Sprintf("%s/%s/json/", c.baseURL, ip)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, fmt.Errorf("new geoip request: %w", err)
	}

	resp, err := c.http.Do(ctx, req)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrLookupFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("%w: unexpected status %d", ErrLookupFailed, resp.StatusCode)
	}

	var payload response
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Result{}, fmt.Errorf("decode geoip response: %w", err)
	}

	if payload.Error {
		msg := payload.Reason
		if msg == "" {
			msg = payload.Message
		}
		if msg == "" {
			msg = "provider returned an error"
		}

		return Result{}, fmt.Errorf("%w: %s", ErrLookupFailed, msg)
	}

	return Result{
		IP:        payload.IP,
		City:      payload.City,
		Country:   payload.CountryName,
		Latitude:  payload.Latitude,
		Longitude: payload.Longitude,
	}, nil
}
