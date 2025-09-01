package semver

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestSemver(t *testing.T) {
	testcases := []struct {
		version1 string
		version2 string
		expected int
	}{
		{
			version1: "1.0.0-alpha",
			version2: "1.0.0-alpha.1",
			expected: -1,
		},
		{
			version1: "1.0.0",
			version2: "1.0.0+1",
			expected: -1,
		},
		{
			version1: "1.0.0-alpha.1",
			version2: "1.0.0-alpha.beta",
			expected: -1,
		},
		{
			version1: "1.0.0-alpha.beta",
			version2: "1.0.0-beta",
			expected: -1,
		},
		{
			version1: "1.0.0-beta",
			version2: "1.0.0-beta.2",
			expected: -1,
		},
		{
			version1: "1.0.0-beta.2",
			version2: "1.0.0-beta.11",
			expected: -1,
		},
		{
			version1: "1.0.0-beta.11",
			version2: "1.0.0-rc.1",
			expected: -1,
		},
		{
			version1: "1.0.0-rc.1",
			version2: "1.0.0",
			expected: -1,
		},
		{
			version1: "1.10.0",
			version2: "2.0.0",
			expected: -1,
		},
		{
			version1: "1.5.0",
			version2: "1.6.0",
			expected: -1,
		},
		{
			version1: "0.5.0",
			version2: "0.6.0",
			expected: -1,
		},
		{
			version1: "0.5.0",
			version2: "2.0.0",
			expected: -1,
		},
		{
			version1: "999.999.999-0.0.1",
			version2: "999.999.999-0.0.2",
			expected: -1,
		},
		{
			version1: "999.999.998-0.0.1",
			version2: "999.999.999-0.0.1",
			expected: -1,
		},
		{
			version1: "v999.999.998-0.0.1",
			version2: "v999.999.999-0.0.1",
			expected: -1,
		},
		{
			version1: "v999.999.999",
			version2: "v999.999.999",
			expected: 0,
		},
		{
			version1: "v999.999.999",
			version2: "v999.999.998",
			expected: 1,
		},
		{
			version1: "1.0.0-alpha+21AF26D3----117B344092BD",
			version2: "1.0.0+21AF26D3----117B344092BD",
			expected: -1,
		},
		{
			version1: "1.0.0",
			version2: "1.0.0+21AF26D3----117B344092BD",
			expected: -1,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.version1+" < "+tc.version2, func(t *testing.T) {
			v1, err := New(tc.version1)
			testutils.Equal(t, err, nil)
			v2, err := New(tc.version2)
			testutils.Equal(t, err, nil)

			bv1, err := v1.Num()
			testutils.Equal(t, err, nil)

			bv2, err := v2.Num()
			testutils.Equal(t, err, nil)

			testutils.Equal(t, bv1.Cmp(bv2), tc.expected)
		})
	}
}

func TestHex(t *testing.T) {
	testcases := []struct {
		version string
		expect  string
	}{
		{
			version: "0.0.0",
			expect:  "000000000000ffffffffffffffff0000",
		},
		{
			version: "0.0.1",
			expect:  "000000000001ffffffffffffffff0000",
		},
		{
			version: "0.0.2-alpha",
			expect:  "000000000002fffb0000000000000000",
		},
		{
			version: "0.0.2-alpha.1",
			expect:  "000000000002fffb0001000000000000",
		},
		{
			version: "0.0.2-alpha.1+21AF26D3----117B344092BD",
			expect:  "000000000002fffb000100000000ffff",
		},
		{
			version: "0.0.2",
			expect:  "000000000002ffffffffffffffff0000",
		},
		{
			version: "0.5.2",
			expect:  "000000050002ffffffffffffffff0000",
		},
		{
			version: "1.0.0",
			expect:  "000100000000ffffffffffffffff0000",
		},
		{
			version: "1.0.0-rc",
			expect:  "000100000000fffe0000000000000000",
		},
		{
			version: "1.0.0+b",
			expect:  "000100000000ffffffffffffffffffff",
		},
		{
			version: "1.1.0",
			expect:  "000100010000ffffffffffffffff0000",
		},
		{
			version: "1.5.0",
			expect:  "000100050000ffffffffffffffff0000",
		},
		{
			version: "2.0.0",
			expect:  "000200000000ffffffffffffffff0000",
		},
		{
			version: "2.3.0",
			expect:  "000200030000ffffffffffffffff0000",
		},
		{
			version: "2.8.0",
			expect:  "000200080000ffffffffffffffff0000",
		},
		{
			version: "3.1.2",
			expect:  "000300010002ffffffffffffffff0000",
		},
		{
			version: "v999.999.999",
			expect:  "03e703e703e7ffffffffffffffff0000",
		},
		{
			version: "v65535.65535.65535-rc",
			expect:  "fffffffffffffffe0000000000000000",
		},
		{
			version: "v65535.65535.65535",
			expect:  "ffffffffffffffffffffffffffff0000",
		},
		{
			version: "v65535.65535.65535+1",
			expect:  "ffffffffffffffffffffffffffffffff",
		},
		{
			version: "v1.0.0",
			expect:  "000100000000ffffffffffffffff0000",
		},
		{
			version: "v1.0.0+1",
			expect:  "000100000000ffffffffffffffffffff",
		},
	}

	for i, tc := range testcases {
		i++

		t.Run(tc.version, func(t *testing.T) {
			vs, err := New(tc.version)
			testutils.Equal(t, err, nil)

			h, err := vs.Hex()
			testutils.Equal(t, err, nil)

			testutils.Equal(t, h, tc.expect)
		})
	}
}
