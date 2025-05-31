package proxy

import (
	"github.com/InsideGallery/core/utils"
)

// Balancer is responsible to control right balance between services
type Balancer struct {
	storage Storage
}

func NewBalancer(s Storage) *Balancer {
	return &Balancer{
		storage: s,
	}
}

func (b *Balancer) AddInstance(subject, instanceID string) error {
	return b.storage.Add(subject, instanceID)
}

func (b *Balancer) RemoveInstance(subject, instanceID string) error {
	return b.storage.Delete(subject, instanceID)
}

func (b *Balancer) DestroyInstance(instanceID string) error {
	return b.storage.DeleteByID(instanceID)
}

func (b *Balancer) GetAllInstances() []string {
	return b.storage.GetIDs()
}

func (b *Balancer) GetSubjectInstances(subject string) []string {
	return b.storage.GetKeys(subject)
}

// Execute takes id and return instanceID for given id
func (b *Balancer) Execute(subject, id string) (string, error) {
	hash := utils.CRC32(id) // We should deliver same hash to same servers

	instances := b.GetSubjectInstances(subject) // This is all instances for given subject
	if len(instances) == 0 {
		return "", ErrNoAvailableInstance
	}

	bucket := int(hash % uint32(len(instances))) // nolint:gosec
	if len(instances) < bucket {
		return "", ErrChooseWrongBucket
	}

	return instances[bucket], nil
}
