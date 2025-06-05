//go:build unit
// +build unit

package mongodb

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/InsideGallery/core/testutils"
)

func TestOID(t *testing.T) {
	tm, err := time.Parse("2006-01-02 15:04:05.999999999 MST", "2022-11-15 00:00:00.560 UTC")
	testutils.Equal(t, err, nil)

	id := primitive.NewObjectIDFromTimestamp(tm)
	fmt.Println(id.Hex())

	tm2, err := time.Parse("2006-01-02 15:04:05.999999999 MST", "2022-11-18 00:00:00.567 UTC")
	testutils.Equal(t, err, nil)

	id2 := primitive.NewObjectIDFromTimestamp(tm2)
	fmt.Println(id2.Hex())
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
	_, err := url.Parse("mongodb://xxx:yyy^@zzz:27017/?replicaSet=rs0&readPreference=secondaryPreferred&retryWrites=false")
	testutils.Equal(t, strings.Contains(err.Error(), "net/url: invalid userinfo"), true)
}
