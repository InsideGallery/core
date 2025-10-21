package semver

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNumericOverflow(t *testing.T) {
	testcases := []struct {
		version string
		desc    string
	}{
		{"65536.0.0", "Major version overflow"},
		{"0.65536.0", "Minor version overflow"},
		{"0.0.65536", "Patch version overflow"},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := New(tc.version)

			require.Error(t, err, ErrVersionIsOverflow)
		})
	}
}

func TestMalformedVersions(t *testing.T) {
	testcases := []struct {
		version string
		reason  string
	}{
		{"1..0", "Empty minor version"},
		{".1.0.0", "Leading dot"},
		{"1.0.0-", "Trailing pre-release delimiter"},
		{"1.0.0+", "Trailing build delimiter"},
		{"1.0.0-+build", "Invalid delimiter order"},
		{"1.0.0-0", "Leading zero"},
	}

	for _, tc := range testcases {
		t.Run(tc.reason, func(t *testing.T) {
			v, err := New(tc.version)
			if err == nil {
				t.Errorf("FAIL: '%s' accepted (%s), hex: %s",
					tc.version, tc.reason, v.Hex())
			}
		})
	}
}

func ExampleSemVersion_Num() {
	vs3, err := New("3")
	if err != nil {
		// that means we are not able to parse version, for example
	}

	vs2, err := New("2")
	if err != nil {
		// that means we are not able to parse version, for example
	}

	fmt.Println(vs3.Num().Cmp(vs2.Num()))
	// Output: 1
}

func ExampleSemVersion_Hex() {
	vs3, err := New("3")
	if err != nil {
		// that means we are not able to parse version
	}

	fmt.Println(vs3.Hex())
	// Output: 000300000000ffff0000000000000000
}

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
			expected: 0,
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
			version1: "1.0.0-rc.1",
			version2: "1.0.0-rc.2",
			expected: -1,
		},
		{
			version1: "1.0.0-1",
			version2: "1.0.0-alpha",
			expected: -1,
		},
		{
			version1: "1.0.0-1",
			version2: "1.0.0-a",
			expected: -1,
		},
		{
			version1: "1.0.0-a",
			version2: "1.0.0-alpha",
			expected: -1,
		},
		{
			version1: "1.0.0-rc-1",
			version2: "1.0.0-rc-2",
			expected: -1,
		},
		{
			version1: "1.0.0-x-z-1",
			version2: "1.0.0-x-z-2",
			expected: -1,
		},
		{
			version1: "1.0.0-rc.1",
			version2: "1.0.0-rc",
			expected: 1,
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
			expected: 0,
		},
		{
			version1: "v2.9",
			version2: "v3",
			expected: -1,
		},
		{
			version1: "v1.0.0",
			version2: "v1",
			expected: 0,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.version1+" < "+tc.version2, func(t *testing.T) {
			v1, err := New(tc.version1)
			assert.Equal(t, err, nil)

			v2, err := New(tc.version2)
			assert.Equal(t, err, nil)

			bv1 := v1.Num()
			bv2 := v2.Num()

			assert.Equal(t, bv1.Cmp(bv2), tc.expected)
		})
	}
}

func TestHex(t *testing.T) {
	testcases := []struct {
		version string
		expect  string
		err     error
	}{
		{
			version: "0.0.0",
			expect:  "000000000000ffff0000000000000000",
		},
		{
			version: "0.0.1",
			expect:  "000000000001ffff0000000000000000",
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
			expect:  "000000000002fffb0001000000000000",
		},
		{
			version: "0.0.2",
			expect:  "000000000002ffff0000000000000000",
		},
		{
			version: "0.5.2",
			expect:  "000000050002ffff0000000000000000",
		},
		{
			version: "1.0.0",
			expect:  "000100000000ffff0000000000000000",
		},
		{
			version: "1.0.0-rc",
			expect:  "000100000000fffe0000000000000000",
		},
		{
			version: "1.0.0+b",
			expect:  "000100000000ffff0000000000000000",
		},
		{
			version: "1.1.0",
			expect:  "000100010000ffff0000000000000000",
		},
		{
			version: "1.5.0",
			expect:  "000100050000ffff0000000000000000",
		},
		{
			version: "2.0.0",
			expect:  "000200000000ffff0000000000000000",
		},
		{
			version: "2.3.0",
			expect:  "000200030000ffff0000000000000000",
		},
		{
			version: "2.8.0",
			expect:  "000200080000ffff0000000000000000",
		},
		{
			version: "3.1.2",
			expect:  "000300010002ffff0000000000000000",
		},
		{
			version: "v999.999.999",
			expect:  "03e703e703e7ffff0000000000000000",
		},
		{
			version: "v65535.65535.65535-rc",
			expect:  "fffffffffffffffe0000000000000000",
		},
		{
			version: "v65535.65535.65535",
			expect:  "ffffffffffffffff0000000000000000",
		},
		{
			version: "v65535.65535.65535+1",
			expect:  "ffffffffffffffff0000000000000000",
		},
		{
			version: "v1.0.0",
			expect:  "000100000000ffff0000000000000000",
		},
		{
			version: "v1.0.0+1",
			expect:  "000100000000ffff0000000000000000",
		},
		{
			version: "3",
			expect:  "000300000000ffff0000000000000000",
		},
		{
			version: "v3",
			expect:  "000300000000ffff0000000000000000",
		},
		{
			version: "v3-1",
			expect:  "00030000000000010000000000000000",
		},
		{
			version: "v3-2",
			expect:  "00030000000000020000000000000000",
		},
		{
			version: "v3-rc",
			expect:  "000300000000fffe0000000000000000",
		},
		{
			version: "test",
			expect:  "000300000000fffe0000000000000000",
			err:     ErrBuildSemver,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.version, func(t *testing.T) {
			vs, err := New(tc.version)
			if tc.err != nil {
				assert.Equal(t, errors.Is(err, tc.err), true)
				return
			}

			assert.Equal(t, err, nil)

			h := vs.Hex()

			assert.Equal(t, h, tc.expect)
		})
	}
}
