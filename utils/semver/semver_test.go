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
