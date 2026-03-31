package location

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"winx-profile/internal/app/core/helpers/errorhandler"
	"winx-profile/internal/app/core/helpers/response"
	resources "winx-profile/internal/app/domain/resources/location"
	service "winx-profile/internal/app/domain/services/location"
	"winx-profile/pkg/geoip"

	"github.com/gin-gonic/gin"
)

var ErrInvalidIPParam = errors.New("invalid ip query param")

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) LookupByIP(ctx *gin.Context) {
	ip, err := clientIP(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(err.Error()))
		errorhandler.FailOnError(err, "resolve client ip")

		return
	}

	location, err := h.service.Lookup(ctx.Request.Context(), ip)
	if err != nil {
		if errors.Is(err, geoip.ErrLookupFailed) {
			ctx.JSON(http.StatusBadGateway, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "LookupGeoIP service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewResource(location), response.OK))
}

func clientIP(ctx *gin.Context) (string, error) {
	if ip := strings.TrimSpace(ctx.Query("ip")); ip != "" {
		if parsed := net.ParseIP(ip); parsed != nil {
			return ip, nil
		}

		return "", ErrInvalidIPParam
	}

	candidates := []string{
		ctx.GetHeader("X-Forwarded-For"),
		ctx.GetHeader("X-Real-IP"),
		ctx.ClientIP(),
		ctx.Request.RemoteAddr,
	}

	for _, candidate := range candidates {
		for _, raw := range strings.Split(candidate, ",") {
			ip := strings.TrimSpace(raw)
			if ip == "" {
				continue
			}

			if host, _, err := net.SplitHostPort(ip); err == nil {
				ip = host
			}

			parsed := net.ParseIP(ip)
			if parsed == nil || !isPublicIP(parsed) {
				continue
			}

			return ip, nil
		}
	}

	return "", nil
}

func isPublicIP(ip net.IP) bool {
	return ip != nil &&
		!ip.IsLoopback() &&
		!ip.IsPrivate() &&
		!ip.IsMulticast() &&
		!ip.IsUnspecified() &&
		!ip.IsLinkLocalMulticast() &&
		!ip.IsLinkLocalUnicast()
}
