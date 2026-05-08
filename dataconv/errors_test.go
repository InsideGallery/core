package dataconv

import (
	"errors"
	"testing"
)

func TestErrorValues(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{"ErrWrongEncodeType_message", ErrWrongEncodeType, "wrong encode type"},
		{"ErrWrongDecodeType_message", ErrWrongDecodeType, "wrong decode type"},
		{"ErrInvalidIPAddress_message", ErrInvalidIPAddress, "invalid ip address"},
		{"ErrNotIPv4Address_message", ErrNotIPv4Address, "not an IPv4 address"},
		{"ErrNotIPv6Address_message", ErrNotIPv6Address, "not an IPv6 address"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatal("expected non-nil error")
			}

			if tc.err.Error() != tc.wantMsg {
				t.Fatalf("expected %q, got %q", tc.wantMsg, tc.err.Error())
			}
		})
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	cases := []struct {
		name string
		a    error
		b    error
	}{
		{"encode_vs_decode", ErrWrongEncodeType, ErrWrongDecodeType},
		{"encode_vs_invalid_ip", ErrWrongEncodeType, ErrInvalidIPAddress},
		{"decode_vs_not_ipv4", ErrWrongDecodeType, ErrNotIPv4Address},
		{"not_ipv4_vs_not_ipv6", ErrNotIPv4Address, ErrNotIPv6Address},
		{"invalid_ip_vs_not_ipv6", ErrInvalidIPAddress, ErrNotIPv6Address},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if errors.Is(tc.a, tc.b) {
				t.Fatalf("expected errors to be distinct: %v and %v", tc.a, tc.b)
			}
		})
	}
}

func TestErrorsIsMatch(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		target error
	}{
		{"ErrWrongEncodeType_self", ErrWrongEncodeType, ErrWrongEncodeType},
		{"ErrWrongDecodeType_self", ErrWrongDecodeType, ErrWrongDecodeType},
		{"ErrInvalidIPAddress_self", ErrInvalidIPAddress, ErrInvalidIPAddress},
		{"ErrNotIPv4Address_self", ErrNotIPv4Address, ErrNotIPv4Address},
		{"ErrNotIPv6Address_self", ErrNotIPv6Address, ErrNotIPv6Address},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !errors.Is(tc.err, tc.target) {
				t.Fatalf("expected errors.Is(%v, %v) to be true", tc.err, tc.target)
			}
		})
	}
}
