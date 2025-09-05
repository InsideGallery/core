package request

import (
	"errors"
	"log/slog"
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var (
	cidrs []*net.IPNet

	ErrAddressIsNotValid = errors.New("address is not valid")
)

func init() {
	maxCidrBlocks := []string{
		"127.0.0.1/8",    // localhost
		"10.0.0.0/8",     // 24-bit block
		"172.16.0.0/12",  // 20-bit block
		"192.168.0.0/16", // 16-bit block
		"169.254.0.0/16", // link local address
		"::1/128",        // localhost IPv6
		"fc00::/7",       // unique local address IPv6
		"fe80::/10",      // link local address IPv6
	}

	cidrs = make([]*net.IPNet, len(maxCidrBlocks))

	for i, maxCidrBlock := range maxCidrBlocks {
		_, cidr, err := net.ParseCIDR(maxCidrBlock)
		if err != nil {
			slog.Default().Error("Error parsing cidr", "err", err)
		}

		cidrs[i] = cidr
	}
}

// IsPrivateAddress check if IP is local IP
func IsPrivateAddress(address string) (bool, error) {
	ipAddress := net.ParseIP(address)
	if ipAddress == nil {
		return false, ErrAddressIsNotValid
	}

	for i := range cidrs {
		if cidrs[i].Contains(ipAddress) {
			return true, nil
		}
	}

	return false, nil
}

// IPFromRequest return ip from request
func IPFromRequest(c *fiber.Ctx) (net.IP, error) {
	var err error

	xRealIP := c.Get("X-Real-Ip")
	xForwardedFor := c.Get("X-Forwarded-For")

	if xRealIP == "" && xForwardedFor == "" {
		remoteIP := c.IP()
		if strings.ContainsRune(remoteIP, ':') {
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
