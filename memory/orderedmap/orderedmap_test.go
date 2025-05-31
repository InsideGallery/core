package orderedmap

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestOrderedmap(t *testing.T) {
	o := &OrderedMap[string, string]{}
	o.Add("A", "1")
	o.Add("B", "2")
	o.Add("C", "3")
	o.Add("E", "4")
	o.Remove("C")
	testutils.Equal(t, o.Get("A"), "1")
	testutils.Equal(t, o.Get("C"), "")
	keys, values := o.GetAll()
	testutils.Equal(t, len(keys), 3)
	testutils.Equal(t, len(values), 3)
	testutils.Equal(t, keys, []string{"A", "B", "E"})
	testutils.Equal(t, values, []string{"1", "2", "4"})
}
