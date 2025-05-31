package entity

import (
	"github.com/InsideGallery/core/db/aerospike"
	as "github.com/aerospike/aerospike-client-go/v7"
)

type Operation struct {
	client     aerospike.Namespace
	pk         interface{}
	setName    string
	expiration uint32
	sendKey    bool
}

func NewOperation(
	client aerospike.Namespace,
	setName string,
	pk interface{},
	sendKey bool,
	expiration uint32,
) *Operation {
	return &Operation{
		client:     client,
		setName:    setName,
		pk:         pk,
		sendKey:    sendKey,
		expiration: expiration,
	}
}

func (h *Operation) Execute(operations []*as.Operation) error {
	policy := as.NewWritePolicy(0, 0)
	policy.SendKey = h.sendKey

	if h.expiration > 0 {
		policy.Expiration = h.expiration
	}

	_, err := h.client.Operate(policy, h.setName, h.pk, operations...)

	return err
}

func (h *Operation) GetNamespace() aerospike.Namespace {
	return h.client
}

func (h *Operation) Get(bins ...string) (*as.Record, error) {
	return h.GetNamespace().Get(nil, h.setName, h.pk, bins...)
}

func (h *Operation) GetBin(binName string) (interface{}, error) {
	rec, err := h.Get(binName)
	if err != nil {
		return nil, err
	}

	if rec == nil {
		return nil, ErrAttributeNotFound
	}

	if rec.Bins == nil {
		return nil, ErrAttributeNotFound
	}

	return rec.Bins[binName], nil
}

func (h *Operation) Exists() (bool, as.Error) {
	return h.GetNamespace().Exists(as.NewPolicy(), h.setName, h.pk)
}
