package storage

import (
	"log/slog"
	"strings"

	aero "github.com/aerospike/aerospike-client-go/v7"

	"github.com/InsideGallery/core/memory/set"
)

type Aerospike struct {
	client    *aero.Client
	namespace string
	setName   string
}

func NewAerospike(namespace string, setName string, client *aero.Client) *Aerospike {
	return &Aerospike{
		client:    client,
		namespace: namespace,
		setName:   setName,
	}
}

func (m *Aerospike) GetKey(group string, id string) string {
	return strings.Join([]string{group, id}, ":")
}

func (m *Aerospike) Add(group string, id string) error {
	key, err := aero.NewKey(m.namespace, m.setName, m.GetKey(group, id))
	if err != nil {
		return err
	}

	return m.client.PutBins(nil, key,
		aero.NewBin("id", id),
		aero.NewBin("group", group),
	)
}

func (m *Aerospike) Delete(group string, id string) error {
	key, err := aero.NewKey(m.namespace, m.setName, m.GetKey(group, id))
	if err != nil {
		return err
	}

	_, err = m.client.Delete(nil, key)

	return err
}

func (m *Aerospike) DeleteByID(id string) error {
	p := aero.NewQueryPolicy()
	p.FilterExpression = aero.ExpEq(aero.ExpStringBin("id"), aero.ExpStringVal(id))

	stmt := aero.NewStatement(m.namespace, m.setName)
	// err := stmt.SetFilter(aero.NewEqualFilter("id", id)) // need secondary index
	// if err != nil {
	// 	return err
	// }

	recordSet, err := m.client.Query(p, stmt)
	if err != nil {
		return err
	}

	var keys []*aero.Key

	for s := range recordSet.Results() {
		group := s.Record.Bins["group"].(string)

		key, err := aero.NewKey(m.namespace, m.setName, m.GetKey(group, id))
		if err != nil {
			return err
		}

		keys = append(keys, key)
	}

	_, err = m.client.BatchDelete(nil, nil, keys)

	return err
}

func (m *Aerospike) GetKeys(group string) []string {
	s := set.NewGenericDataSet[string]()

	p := aero.NewQueryPolicy()
	p.FilterExpression = aero.ExpEq(aero.ExpStringBin("group"), aero.ExpStringVal(group))

	stmt := aero.NewStatement(m.namespace, m.setName)
	// err := stmt.SetFilter(aero.NewEqualFilter("group", group)) // need secondary index
	// if err != nil {
	// 	return err
	// }

	recordSet, err := m.client.Query(p, stmt)
	if err != nil {
		slog.Error("Error while querying aerospike:", "err", err)
		return nil
	}

	for rs := range recordSet.Results() {
		// group := s.Record.Bins["group"].(string)
		id := rs.Record.Bins["id"].(string)

		s.Add(id)
	}

	return s.ToSlice()
}

func (m *Aerospike) GetIDs() []string {
	s := set.NewGenericDataSet[string]()
	p := aero.NewQueryPolicy()
	stmt := aero.NewStatement(m.namespace, m.setName)

	recordSet, err := m.client.Query(p, stmt)
	if err != nil {
		slog.Error("Error while querying aerospike:", "err", err)
		return nil
	}

	for rs := range recordSet.Results() {
		// group := s.Record.Bins["group"].(string)
		id := rs.Record.Bins["id"].(string)

		s.Add(id)
	}

	return s.ToSlice()
}

func (m *Aerospike) Size(group string) int {
	p := aero.NewQueryPolicy()
	p.FilterExpression = aero.ExpEq(aero.ExpStringBin("group"), aero.ExpStringVal(group))

	stmt := aero.NewStatement(m.namespace, m.setName)
	// err := stmt.SetFilter(aero.NewEqualFilter("group", group)) // need secondary index
	// if err != nil {
	// 	return err
	// }

	recordSet, err := m.client.Query(p, stmt)
	if err != nil {
		slog.Error("Error while querying aerospike:", "err", err)
		return 0
	}

	var count int
	for range recordSet.Results() {
		count++
	}

	return count
}
