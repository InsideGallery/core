package mongodb

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/InsideGallery/core/testutils"
)

func TestOID(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{name: "first_timestamp", raw: "2022-11-15 00:00:00.560 UTC"},
		{name: "second_timestamp", raw: "2022-11-18 00:00:00.567 UTC"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tm, err := time.Parse("2006-01-02 15:04:05.999999999 MST", tc.raw)
			testutils.Equal(t, err, nil)

			id := primitive.NewObjectIDFromTimestamp(tm)
			testutils.Equal(t, id.Timestamp(), tm.Truncate(time.Second))
		})
	}
}

func TestNewMongoClient(t *testing.T) {
	config := &ConnectionConfig{
		Hosts:    []string{"localhost:27017"},
		User:     "xxxx",
		Scheme:   "mongodb",
		Pass:     "^",
		Database: "",
		Args:     "replicaSet=rs0&readPreference=secondaryPreferred&retryWrites=false",
	}

	mongoClient, err := NewMongoClient(config)
	testutils.Equal(t, err, nil)
	testutils.Equal(t, mongoClient != nil, true)
}

func TestURL(t *testing.T) {
	rawURL := strings.Join([]string{
		"mongodb://xxx:",
		"yyy^",
		"@zzz:27017/?replicaSet=rs0&readPreference=secondaryPreferred&retryWrites=false",
	}, "")

	_, err := url.Parse(rawURL)
	testutils.Equal(t, strings.Contains(err.Error(), "net/url: invalid userinfo"), true)
}
