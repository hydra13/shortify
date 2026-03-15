package trustsubnet

import (
	"fmt"
	"net"
	"net/http"

	"github.com/rs/zerolog"
)

type TrustedSubnetMiddleware struct {
	trustedSubnet *net.IPNet
	log           zerolog.Logger
}

func NewTrustedSubnetMiddleware(cidr string, log zerolog.Logger) (func(http.Handler) http.Handler, error) {
	if cidr == "" {
		// Возвращаем middleware который всегда возвращает 403
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			})
		}, nil
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}

	middleware := &TrustedSubnetMiddleware{
		trustedSubnet: ipNet,
		log:           log,
	}

	return middleware.Middleware, nil
}

func (m *TrustedSubnetMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipStr := r.Header.Get("X-Real-IP")
		if ipStr == "" {
			m.log.Debug().Msg("X-Real-IP header is missing")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		ip := net.ParseIP(ipStr)
		if ip == nil {
			m.log.Debug().Str("ip", ipStr).Msg("invalid IP address")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if !m.trustedSubnet.Contains(ip) {
			m.log.Debug().
				Str("client_ip", ipStr).
				Str("trusted_subnet", m.trustedSubnet.String()).
				Msg("IP is not in trusted subnet")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
