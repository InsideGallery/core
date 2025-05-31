//go:build unit
// +build unit

package utils

import (
	"encoding/json"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestPassword(t *testing.T) {
	p := Password("secret")
	testutils.Equal(t, p.String(), "********")
	b, err := json.Marshal(map[string]interface{}{
		"password": p,
	})
	testutils.Equal(t, err, nil)
	testutils.Equal(t, string(b), `{"password":"********"}`)
	testutils.Equal(t, p.Value(), "secret")
}
