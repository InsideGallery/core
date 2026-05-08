package dataconv

import (
	"math/big"
	"net"
	"testing"
)

func TestIPV4ToIPV6TableDriven(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"standard_ip", "192.168.1.1", "::ffff:192.168.1.1"},
		{"loopback", "127.0.0.1", "::ffff:127.0.0.1"},
		{"all_zeros", "0.0.0.0", "::ffff:0.0.0.0"},
		{"broadcast", "255.255.255.255", "::ffff:255.255.255.255"},
		{"empty_string", "", "::ffff:"},
		{"dns_google", "8.8.8.8", "::ffff:8.8.8.8"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IPV4ToIPV6(tc.in)
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestIPV6ToStringTableDriven(t *testing.T) {
	cases := []struct {
		name string
		in   *big.Int
		want string
	}{
		{"zero", big.NewInt(0), "A000000000000000000000000000000000000000"},
		{"one", big.NewInt(1), "A000000000000000000000000000000000000001"},
		{"small_number", big.NewInt(12345), "A000000000000000000000000000000000012345"},
		{"ipv4_as_bigint", big.NewInt(3232235777), "A000000000000000000000000000003232235777"},
		{"large_ipv6", GetBigInt("340282366920938463463374607431768211455"), "A340282366920938463463374607431768211455"},
		{"exact_39_chars", GetBigInt("123456789012345678901234567890123456789"), "A123456789012345678901234567890123456789"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IPV6ToString(tc.in)
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestIP2IntTableDriven(t *testing.T) {
	cases := []struct {
		name string
		in   net.IP
		want *big.Int
	}{
		{"loopback_v4", net.ParseIP("127.0.0.1"), big.NewInt(2130706433)},
		{"zeros_v4", net.ParseIP("0.0.0.0"), big.NewInt(0)},
		{"broadcast_v4", net.ParseIP("255.255.255.255"), big.NewInt(4294967295)},
		{"google_dns", net.ParseIP("8.8.8.8"), big.NewInt(134744072)},
		{"private_v4", net.ParseIP("10.0.0.1"), big.NewInt(167772161)},
		{"ipv6_loopback", net.ParseIP("::1"), big.NewInt(1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IP2Int(tc.in)
			if got.Cmp(tc.want) != 0 {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestIPv4ToIntTableDriven(t *testing.T) {
	cases := []struct {
		name    string
		in      net.IP
		want    uint32
		wantErr bool
	}{
		{"standard", net.ParseIP("192.168.1.1"), 3232235777, false},
		{"zeros", net.ParseIP("0.0.0.0"), 0, false},
		{"broadcast", net.ParseIP("255.255.255.255"), 4294967295, false},
		{"loopback", net.ParseIP("127.0.0.1"), 2130706433, false},
		{"ipv6_not_v4", net.ParseIP("2001:db8::1"), 0, true},
		{"nil_ip", nil, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := IPv4ToInt(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

func TestIPv6ToIntTableDriven(t *testing.T) {
	cases := []struct {
		name string
		in   net.IP
		want [2]uint64
	}{
		{"all_zeros", net.ParseIP("::"), [2]uint64{0, 0}},
		{"loopback", net.ParseIP("::1"), [2]uint64{0, 1}},
		{"google_dns_v6", net.ParseIP("2001:4860:4860::8888"), [2]uint64{2306204062558715904, 34952}},
		{"high_only", net.ParseIP("ffff::"), [2]uint64{18446462598732840960, 0}},
		{"all_ones", net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"), [2]uint64{18446744073709551615, 18446744073709551615}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := IPv6ToInt(tc.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestIPv6ToBigIntTableDriven(t *testing.T) {
	cases := []struct {
		name string
		in   net.IP
		want *big.Int
	}{
		{"all_zeros", net.ParseIP("::"), GetBigInt("0")},
		{"loopback", net.ParseIP("::1"), GetBigInt("1")},
		{"all_ones", net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"), GetBigInt("340282366920938463463374607431768211455")},
		{"google_dns", net.ParseIP("2001:4860:4860::8888"), GetBigInt("42541956123769884636017138956568135816")},
		{"link_local", net.ParseIP("fe80::1"), GetBigInt("338288524927261089654018896841347694593")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IPv6ToBigInt(tc.in)
			if got.Cmp(tc.want) != 0 {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestIntToIPv4TableDriven(t *testing.T) {
	cases := []struct {
		name string
		in   uint32
		want net.IP
	}{
		{"zeros", 0, net.ParseIP("0.0.0.0")},
		{"loopback", 2130706433, net.ParseIP("127.0.0.1")},
		{"broadcast", 4294967295, net.ParseIP("255.255.255.255")},
		{"private", 3232235777, net.ParseIP("192.168.1.1")},
		{"one", 1, net.IPv4(0, 0, 0, 1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IntToIPv4(tc.in)
			if !got.Equal(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestIntToIPv6TableDriven(t *testing.T) {
	cases := []struct {
		name string
		high uint64
		low  uint64
		want net.IP
	}{
		{"all_zeros", 0, 0, net.ParseIP("::")},
		{"loopback", 0, 1, net.ParseIP("::1")},
		{"high_only", 1, 0, net.ParseIP("::1:0:0:0:0")},
		{"all_max", 18446744073709551615, 18446744073709551615, net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")},
		{"google_dns", 2306204062558715904, 34952, net.ParseIP("2001:4860:4860::8888")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IntToIPv6(tc.high, tc.low)
			if !got.Equal(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestBigIntToIPv6TableDriven(t *testing.T) {
	cases := []struct {
		name string
		in   *big.Int
		want net.IP
	}{
		{"zero", GetBigInt("0"), net.ParseIP("::")},
		{"one", GetBigInt("1"), net.ParseIP("::1")},
		{"all_max", GetBigInt("340282366920938463463374607431768211455"), net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")},
		{"known_address", GetBigInt("42540488161975842760550637899214225665"), net.ParseIP("2001:0:0:0:0:ffff:c0a8:101")},
		{"google_dns", GetBigInt("42541956123769884636017138956568135816"), net.ParseIP("2001:4860:4860::8888")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := BigIntToIPv6(*tc.in)
			if !got.Equal(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestParseIPTableDriven(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantLen int
		wantErr bool
	}{
		{"ipv4_standard", "192.168.1.1", net.IPv4len, false},
		{"ipv4_loopback", "127.0.0.1", net.IPv4len, false},
		{"ipv6_standard", "2001:db8::1", net.IPv6len, false},
		{"ipv6_full", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", net.IPv6len, false},
		{"invalid_hostname", "google.com", 0, true},
		{"empty_string", "", 0, true},
		{"garbage", "not-an-ip", 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip, length, err := ParseIP(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if ip != nil {
					t.Fatal("expected nil ip on error")
				}
				if length != 0 {
					t.Fatalf("expected 0 length on error, got %d", length)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if length != tc.wantLen {
				t.Fatalf("expected length %d, got %d", tc.wantLen, length)
			}
			if ip == nil {
				t.Fatal("expected non-nil ip")
			}
		})
	}
}

func TestCutIPTableDriven(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty_string", "", ""},
		{"invalid_ip", "not-an-ip", ""},
		{"ipv4_last_octet_zeroed", "192.168.10.1", "192.168.10.0"},
		{"ipv4_already_zero", "10.0.0.0", "10.0.0.0"},
		{"ipv4_broadcast", "255.255.255.255", "255.255.255.0"},
		{"ipv6_standard", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", "2001:db8:85a3::8a2e:370:7300"},
		{"ipv4_mapped_ipv6", "::ffff:192.0.2.128", "192.0.2.0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CutIP(tc.in)
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestIPv4ToIntRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		ip   string
	}{
		{"loopback", "127.0.0.1"},
		{"private_a", "10.0.0.1"},
		{"private_b", "172.16.0.1"},
		{"private_c", "192.168.1.1"},
		{"public", "8.8.4.4"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			original := net.ParseIP(tc.ip)
			intVal, err := IPv4ToInt(original)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			recovered := IntToIPv4(intVal)
			if !recovered.Equal(original) {
				t.Fatalf("round trip failed: %v -> %d -> %v", original, intVal, recovered)
			}
		})
	}
}

func TestIPv6BigIntRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		ip   string
	}{
		{"loopback", "::1"},
		{"all_zeros", "::"},
		{"google_dns", "2001:4860:4860::8888"},
		{"all_ones", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"},
		{"link_local", "fe80::1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			original := net.ParseIP(tc.ip)
			bigVal := IPv6ToBigInt(original)
			recovered := BigIntToIPv6(*bigVal)
			if !recovered.Equal(original) {
				t.Fatalf("round trip failed: %v -> %v -> %v", original, bigVal, recovered)
			}
		})
	}
}
