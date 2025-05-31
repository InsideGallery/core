//go:build unit
// +build unit

package dataconv

import (
	"testing"

	"dario.cat/mergo"

	"github.com/AlekSi/pointer"

	"github.com/InsideGallery/core/testutils"
)

type Nested struct {
	Field1 string
	Field2 string
}

type InputRecord struct {
	SessionID     *string
	EventID       *string
	RequestID     *string
	AccountLogin  *string
	AccountLogin2 string
	Nested        Nested
}

func TestMergeStruct(t *testing.T) {
	sesID := "ses123"
	reqID := "req123"
	entID := "evt123"
	usrID := "user123"
	rec := &InputRecord{
		SessionID: &sesID,
		EventID:   &entID,
		Nested: Nested{
			Field1: "value1",
		},
	}
	rec2 := &InputRecord{
		RequestID:     &reqID,
		AccountLogin:  &usrID,
		AccountLogin2: usrID,
		Nested: Nested{
			Field2: "value2",
		},
	}
	err := MergeStruct(rec, rec2)
	testutils.Equal(t, err, nil)
	testutils.Equal(t, rec.SessionID, pointer.ToString(sesID))
	testutils.Equal(t, rec.RequestID, rec2.RequestID)
	testutils.Equal(t, rec.AccountLogin, rec2.AccountLogin)
	testutils.Equal(t, rec.AccountLogin2, rec2.AccountLogin2)
	testutils.Equal(t, rec.Nested.Field1, "value1")
	testutils.Equal(t, rec.Nested.Field2, rec2.Nested.Field2)
}

func TestMerge(t *testing.T) {
	sesID := "ses123"
	reqID := "req123"
	entID := "evt123"
	usrID := "user123"
	rec := InputRecord{
		SessionID: &sesID,
		EventID:   &entID,
		Nested: Nested{
			Field1: "value1",
		},
	}
	rec2 := InputRecord{
		RequestID:     &reqID,
		AccountLogin:  &usrID,
		AccountLogin2: usrID,
		Nested: Nested{
			Field2: "value2",
		},
	}

	err := MergeStruct(&rec, rec2)
	testutils.Equal(t, err, nil)

	testutils.Equal(t, rec.SessionID, pointer.ToString(sesID))
	testutils.Equal(t, rec.RequestID, rec2.RequestID)
	testutils.Equal(t, rec.AccountLogin, rec2.AccountLogin)
	testutils.Equal(t, rec.AccountLogin2, rec2.AccountLogin2)
	testutils.Equal(t, rec.Nested.Field1, "value1")
	testutils.Equal(t, rec.Nested.Field2, rec2.Nested.Field2)
}

func TestMergeMaps(t *testing.T) {
	sesID := "ses123"
	reqID := "req123"
	entID := "evt123"
	usrID := "user123"
	rec := map[string]interface{}{
		"session_id": sesID,
		"req_id":     reqID,
		"nested": map[string]interface{}{
			"field1": "value1",
		},
	}
	rec2 := map[string]interface{}{
		"session_id": "abc",
		"event_id":   entID,
		"user_id":    usrID,
		"nested": map[string]interface{}{
			"field2": "value2",
		},
	}

	err := MergeStruct(&rec, rec2)
	testutils.Equal(t, err, nil)
	testutils.Equal(t, rec["event_id"], rec2["event_id"])
	testutils.Equal(t, rec["user_id"], rec2["user_id"])

	nested, ok := rec["nested"].(map[string]interface{})
	testutils.Equal(t, ok, true)
	testutils.Equal(t, nested["field1"], "value1")
	testutils.Equal(t, nested["field2"], "value2")
}

func TestMergeNonPointerStructs(t *testing.T) {
	type testStruct struct {
		Val int
	}

	err := MergeStruct(testStruct{Val: 1}, testStruct{Val: 2})
	testutils.Equal(t, err, mergo.ErrNonPointerArgument)
}
