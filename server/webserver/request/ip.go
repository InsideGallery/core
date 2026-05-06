package request

import (
	"errors"
	"net"
	"strings"

	"github.com/gofiber/fiber/v3"
)

var ErrAddressIsNotValid = errors.New("address is not valid")

// IsPrivateAddress checks whether address is loopback, private, or link-local.
func IsPrivateAddress(address string) (bool, error) {
	ipAddress := net.ParseIP(address)
	if ipAddress == nil {
		return false, ErrAddressIsNotValid
	}

	return ipAddress.IsLoopback() || ipAddress.IsPrivate() || ipAddress.IsLinkLocalUnicast(), nil
}

// IPFromRequest extracts the real client IP from a Fiber request.
func IPFromRequest(c fiber.Ctx) (net.IP, error) {
	xRealIP := c.Get("X-Real-Ip")
	xForwardedFor := c.Get("X-Forwarded-For")

	if xRealIP == "" && xForwardedFor == "" {
		remoteIP := c.IP()
		if strings.ContainsRune(remoteIP, ':') {
			var err error

			remoteIP, _, err = net.SplitHostPort(remoteIP)
			if err != nil {
				return nil, err
			}
		}

		return net.ParseIP(remoteIP), nil
	}

	for _, address := range strings.Split(xForwardedFor, ",") {
		address = strings.TrimSpace(address)
		isPrivate, err := IsPrivateAddress(address)

		if !isPrivate && err == nil {
			return net.ParseIP(address), nil
		}
	}

	return net.ParseIP(xRealIP), nil
}

// IPStringFromRequest returns the request IP as a string.
func IPStringFromRequest(c fiber.Ctx) string {
	ip, err := IPFromRequest(c)
	if err != nil || ip == nil {
		return c.IP()
	}

	return ip.String()
}
