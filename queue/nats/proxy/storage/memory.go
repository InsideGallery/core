package storage

import (
	"github.com/InsideGallery/core/memory/registry"
	"github.com/InsideGallery/core/memory/set"
)

type Memory struct {
	registry *registry.Registry[string, string, any]
}

func NewMemory() *Memory {
	return &Memory{
		registry: registry.NewRegistry[string, string, any](),
	}
}

func (m *Memory) Add(group string, id string) error {
	return m.registry.Add(group, id, registry.Nothing)
}

func (m *Memory) Delete(group string, id string) error {
	return m.registry.Remove(group, id)
}

func (m *Memory) DeleteByID(id string) error {
	return m.registry.RemoveIDEverywhere(id)
}

func (m *Memory) GetKeys(group string) []string {
	return m.registry.GetGroup(group).GetKeys()
}

func (m *Memory) GetIDs() []string {
	s := set.NewGenericDataSet[string]()
	subjects := m.registry.GetKeys()

	for _, groupID := range subjects {
		group := m.registry.GetGroup(groupID)
		for _, id := range group.GetKeys() {
			s.Add(id)
		}
	}

	return s.ToSlice()
}

func (m *Memory) Size(group string) int {
	return m.registry.GetGroup(group).Size()
}
