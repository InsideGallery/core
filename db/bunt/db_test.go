package bunt

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

type Test struct {
	Name string
}

func TestDB(t *testing.T) {
	w, err := GetConnection()
	if err != nil {
		t.Fatal(err)
	}

	name := "val"
	ts := &Test{
		Name: name,
	}

	err = w.Set("key1", ts)
	if err != nil {
		t.Fatal(err)
	}

	ts = new(Test)

	err = w.Get("key1", ts)
	if err != nil {
		t.Fatal(err)
	}

	testutils.Equal(t, ts.Name, name)
}
