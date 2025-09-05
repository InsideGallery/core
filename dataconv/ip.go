package dataconv

import (
	"encoding/binary"
	"errors"
	"math/big"
	"net"
	"strings"
)

const (
	ArrayLenForIP = 8
	LowPosition   = 8
	HighPosition  = 16

	indexFillSize   = 39
	indexFillSymbol = "0"
	indexMarker     = "A"

	IPV6Prefix = "::ffff:"

	IPBytesCut = 15 // 15 bytes (120 bits)
)

var (
	ErrInvalidIPAddress = errors.New("invalid ip address")
	ErrNotIPv4Address   = errors.New("not an IPv4 address")
	ErrNotIPv6Address   = errors.New("not an IPv6 address")
)

func IPV4ToIPV6(ipv4 string) string { //nolint:revive
	return strings.Join([]string{IPV6Prefix, ipv4}, "")
}

func IPV6ToString(ipv6 *big.Int) string { //nolint:revive
	var zeroes string

	rawip := ipv6.String()

	size := len(rawip)
	if indexFillSize-size > 0 {
		zeroes = strings.Repeat(indexFillSymbol, indexFillSize-size)
	}

	return strings.Join([]string{indexMarker, zeroes, rawip}, "")
}

// IP2Int convert any net.IP to uint64
func IP2Int(ip net.IP) *big.Int { //nolint:revive
	ip4 := ip.To4()
	if ip4 != nil {
		i := big.NewInt(0)
		i.SetBytes(ip4)

		return i
	}

	i := big.NewInt(0)
	i.SetBytes(ip)

	return i
}

// IPv4ToInt converts IP address of version 4 from net.IP to uint32
// representation.
func IPv4ToInt(ipaddr net.IP) (uint32, error) { //nolint:revive
	if ipaddr.To4() == nil {
		return 0, ErrNotIPv4Address
	}

	return binary.BigEndian.Uint32(ipaddr.To4()), nil
}

// IPv6ToInt converts IP address of version 6 from net.IP to uint64 array
// representation. Return value contains high integer value on the first
// place and low integer value on second place.
func IPv6ToInt(ipaddr net.IP) ([2]uint64, error) { //nolint:revive
	if ipaddr.To16()[0:LowPosition] == nil || ipaddr.To16()[LowPosition:HighPosition] == nil {
		return [2]uint64{0, 0}, ErrNotIPv6Address
	}

	// Get two separates values of integer IP
	ip := [2]uint64{
		binary.BigEndian.Uint64(ipaddr.To16()[0:LowPosition]),            // IP high
		binary.BigEndian.Uint64(ipaddr.To16()[LowPosition:HighPosition]), // IP low
	}

	return ip, nil
}

// IPv6ToBigInt converts IP address of version 6 from net.IP to math big
// integer representation.
func IPv6ToBigInt(ipaddr net.IP) *big.Int { //nolint:revive
	// Initialize value as bytes
	var ip big.Int
	ip.SetBytes(ipaddr)

	return &ip
}

// IntToIPv4 converts IP address of version 4 from integer to net.IP
// representation.
func IntToIPv4(ipaddr uint32) net.IP {
	ip := make(net.IP, net.IPv4len)

	// Proceed conversion
	binary.BigEndian.PutUint32(ip, ipaddr)

	return ip
}

// IntToIPv6 converts IP address of version 6 from integer (high and low value)
// to net.IP representation.
func IntToIPv6(high, low uint64) net.IP {
	ip := make(net.IP, net.IPv6len)

	// Allocate 8 bytes arrays for IPs
	ipHigh := make([]byte, ArrayLenForIP)
	ipLow := make([]byte, ArrayLenForIP)

	// Proceed conversion
	binary.BigEndian.PutUint64(ipHigh, high)
	binary.BigEndian.PutUint64(ipLow, low)

	for i := 0; i < net.IPv6len; i++ {
		if i < ArrayLenForIP {
			ip[i] = ipHigh[i]
		} else if i >= ArrayLenForIP {
			ip[i] = ipLow[i-ArrayLenForIP]
		}
	}

	return ip
}

// BigIntToIPv6 converts IP address of version 6 from big integer to net.IP
// representation.
func BigIntToIPv6(ipaddr big.Int) net.IP {
	ip := make(net.IP, net.IPv6len)

	ipBytes := ipaddr.Bytes()
	ipBytesLen := len(ipBytes)

	for i := 0; i < net.IPv6len; i++ {
		if i < net.IPv6len-ipBytesLen {
			ip[i] = 0x0
		} else {
			ip[i] = ipBytes[ipBytesLen-net.IPv6len+i]
		}
	}

	return ip
}

// ParseIP implements extension of net.ParseIP. It returns additional
// information about IP address bytes length. In general, it works typically
// as standard net.ParseIP. So if IP is not valid, nil is returned.
func ParseIP(s string) (net.IP, int, error) {
	pip := net.ParseIP(s)
	if pip == nil {
		return nil, 0, ErrInvalidIPAddress
	} else if strings.Contains(s, ".") {
		return pip, net.IPv4len, nil
	}

	return pip, net.IPv6len, nil
}

func CutIP(ip string) string {
	if ip == "" {
		return ""
	}

	parsedIP := net.ParseIP(ip)
	if len(parsedIP) < IPBytesCut {
		return ""
	}

	parsedIP[IPBytesCut] = 0

	return parsedIP.String()
}
