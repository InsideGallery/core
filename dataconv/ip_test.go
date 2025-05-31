//go:build unit
// +build unit

package dataconv

import (
	"math/big"
	"net"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestIPV6ToString(t *testing.T) {
	rawip := "46.219.132.112"
	rip := net.ParseIP(rawip)
	intip := IPv6ToBigInt(rip)
	index := IPV6ToString(intip)
	testutils.Equal(t, index, "A000000000000000000000000281471467881584")
}

func TestIPV6ToString1(t *testing.T) {
	ipv4 := "46.219.132.112"
	ipv6 := IPV4ToIPV6(ipv4)
	testutils.Equal(t, ipv6, "::ffff:46.219.132.112")
}

func TestIP2Int(t *testing.T) {
	ipv6 := GetBigInt("42541956123769884636017138956568135816")

	for _, c := range []struct {
		in   net.IP
		want *big.Int
	}{
		{net.ParseIP("192.168.1.1"), big.NewInt(3232235777)},
		{net.ParseIP("0.0.0.0"), big.NewInt(0)},
		{net.ParseIP("8.8.8.8"), big.NewInt(134744072)},
		{net.ParseIP("255.255.255.255"), big.NewInt(4294967295)},
		{net.ParseIP("20.36.77.12"), big.NewInt(337923340)},
		{net.ParseIP("2001:4860:4860::8888"), ipv6},
	} {
		got := IP2Int(c.in)
		if got.Cmp(c.want) != 0 {
			t.Errorf("Ip2Int(%v) == %v, want %v", c.in, got, c.want)
		}
	}
}

func TestIPv4ToInt(t *testing.T) {
	for _, c := range []struct {
		in   net.IP
		want uint32
	}{
		{net.ParseIP("192.168.1.1"), 3232235777},
		{net.ParseIP("0.0.0.0"), 0},
		{net.ParseIP("8.8.8.8"), 134744072},
		{net.ParseIP("255.255.255.255"), 4294967295},
		{net.ParseIP("20.36.77.12"), 337923340},
	} {
		got, err := IPv4ToInt(c.in)
		if got != c.want || err != nil {
			t.Errorf("IPv4ToInt(%v) == %v, want %v", c.in, got, c.want)
		}
	}
}

func TestIPv4ToIntError(t *testing.T) {
	for _, c := range []struct {
		in   net.IP
		want uint32
	}{
		{net.ParseIP("google.com"), 0},
	} {
		got, err := IPv4ToInt(c.in)
		if err == nil {
			t.Errorf("IPv4ToInt(%v) == %v, want %v", c.in, got, c.want)
		}
	}
}

func TestIPv6ToInt(t *testing.T) {
	for _, c := range []struct {
		in   net.IP
		want [2]uint64
	}{
		{net.ParseIP("0000:0000:0000:0000:0000:0000:0000:0000"), [2]uint64{0, 0}},
		{net.ParseIP("0000:0000:0000:0000:0000:0000:0000:1"), [2]uint64{0, 1}},
		{net.ParseIP("2001:4860:4860::8888"), [2]uint64{2306204062558715904, 34952}},
	} {
		got, err := IPv6ToInt(c.in)
		if got != c.want || err != nil {
			t.Errorf("IPv6ToInt(%v) == %v, want %v", c.in.To16(), got, c.want)
		}
	}
}

func TestIPv6ToBigInt(t *testing.T) {
	for _, c := range []struct {
		in   net.IP
		want *big.Int
	}{
		{net.ParseIP("0000:0000:0000:0000:0000:0000:0000:0000"), GetBigInt("0")},
		{net.ParseIP("0000:0000:0000:0000:0000:0000:0000:0001"), GetBigInt("1")},
		{net.ParseIP("2001:0:0:0:0:ffff:c0a8:101"), GetBigInt("42540488161975842760550637899214225665")},
		{net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"), GetBigInt("340282366920938463463374607431768211455")},
	} {
		got := IPv6ToBigInt(c.in)
		if got.Cmp(c.want) != 0 {
			t.Errorf("IPv6ToInt(%v) == %v, want %v", c.in.To16(), got, c.want)
		}
	}
}

func TestIntToIPv4(t *testing.T) {
	for _, c := range []struct {
		in   uint32
		want net.IP
	}{
		{3232235777, net.ParseIP("192.168.1.1")},
		{0, net.ParseIP("0.0.0.0")},
		{134744072, net.ParseIP("8.8.8.8")},
		{4294967295, net.ParseIP("255.255.255.255")},
	} {
		got := IntToIPv4(c.in)
		if !got.Equal(c.want) {
			t.Errorf("IntToIPv4(%v) == %v, want %v", c.in, got, c.want)
		}
	}
}

func TestIntToIPv6(t *testing.T) {
	for _, c := range []struct {
		in   [2]uint64
		want net.IP
	}{
		{[2]uint64{0, 0}, net.ParseIP("0000:0000:0000:0000:0000:0000:0000:0000")},
		{[2]uint64{0, 1}, net.ParseIP("0000:0000:0000:0000:0000:0000:0000:0001")},
		{[2]uint64{1, 0}, net.ParseIP("0000:0000:0000:0001:0000:0000:0000:0000")},
		{[2]uint64{2306204062558715904, 34952}, net.ParseIP("2001:4860:4860::8888")},
		{[2]uint64{0, 18446744073709551615}, net.ParseIP("0000:0000:0000:0000:ffff:ffff:ffff:ffff")},
		{[2]uint64{18446744073709551615, 18446744073709551615}, net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")},
	} {
		got := IntToIPv6(c.in[0], c.in[1])
		if !got.Equal(c.want) {
			t.Errorf("IntToIPv6(%v) == %v, want %v", c.in, got, c.want)
		}
	}
}

func TestBigIntToIPv6(t *testing.T) {
	for _, c := range []struct {
		in   *big.Int
		want net.IP
	}{
		{GetBigInt("0"), net.ParseIP("0000:0000:0000:0000:0000:0000:0000:0000")},
		{GetBigInt("1"), net.ParseIP("0000:0000:0000:0000:0000:0000:0000:0001")},
		{GetBigInt("42540488161975842760550637899214225665"), net.ParseIP("2001:0:0:0:0:ffff:c0a8:101")},
		{GetBigInt("340282366920938463463374607431768211455"), net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")},
	} {
		got := BigIntToIPv6(*c.in)
		if !got.Equal(c.want) {
			t.Errorf("BigIntToIPv6(%v) == %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseIP(t *testing.T) {
	for _, c := range []struct {
		in   string
		want int
	}{
		{"192.168.1.1", 4},
		{"0.0.0.0", 4},
		{"0000:0000:0000:0000:0000:0000:0000:0000", 16},
		{"0000:0000:0000:0000:0000:0000:0000:0001", 16},
		{"2001:0:0:0:0:ffff:c0a8:101", 16},
	} {
		_, gotType, err := ParseIP(c.in)
		if gotType != c.want || err != nil {
			t.Errorf("ParseIP(%v) == %v, want %v", c.in, gotType, c.want)
		}
	}
}

func TestParseIPError(t *testing.T) {
	for _, c := range []struct {
		in   string
		want int
	}{
		{"google.com", 0},
	} {
		_, gotType, err := ParseIP(c.in)
		if gotType != c.want || err == nil {
			t.Errorf("ParseIP(%v) == %v, want %v", c.in, gotType, c.want)
		}
	}
}

func GetBigInt(bi string) *big.Int {
	bigInt := new(big.Int)
	bigInt.SetString(bi, 10)
	return bigInt
}

func TestCutIP(t *testing.T) {
	tests := []struct {
		name string
		IP   string
		want string
	}{
		{
			name: "empty",
			IP:   "",
			want: "",
		},
		{
			name: "not ip",
			IP:   "bandera",
			want: "",
		},
		{
			name: "ipv4",
			IP:   "192.168.10.1",
			want: "192.168.10.0",
		},
		{
			name: "ipv4",
			IP:   "192.168.0.2",
			want: "192.168.0.0",
		},
		{
			name: "ipv6",
			IP:   "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			want: "2001:db8:85a3::8a2e:370:7300",
		},
		{
			name: "ipv6",
			IP:   "2001:0db8:85a3:0000:0000:8a2e:0370:7335",
			want: "2001:db8:85a3::8a2e:370:7300",
		},
		{
			name: "ipv6",
			IP:   "2001:0db8:85a3:0000:0000:8a2e:0370:7336",
			want: "2001:db8:85a3::8a2e:370:7300",
		},
		{
			name: "IPv6 Expanded",
			IP:   "0000:0000:0000:0000:0000:ffff:c158:632e",
			want: "193.88.99.0",
		},
		{
			name: "IPv6 Compressed",
			IP:   "::ffff:c158:632e",
			want: "193.88.99.0",
		},
		{
			name: "IPv6 Expanded (Shortened)",
			IP:   "0:0:0:0:0:ffff:c158:632e",
			want: "193.88.99.0",
		},
		{
			name: "IPv6 manually",
			IP:   "2002:c158:632e:0001::1",
			want: "2002:c158:632e:1::",
		},
		{
			name: "IPv4 like IPv6",
			IP:   "::ffff:192.0.2.128",
			want: "192.0.2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.Equal(t, CutIP(tt.IP), tt.want)
		})
	}
}
