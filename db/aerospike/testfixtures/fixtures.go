package testfixtures

import (
	"encoding/json"

	"github.com/aerospike/aerospike-client-go/v7"
)

type AerospikeFixture struct {
	Bins      map[string]interface{} `json:"bins"`
	Namespace string                 `json:"namespace"`
	Set       string                 `json:"set"`
	Key       string                 `json:"key"`
}

// LoadAerospikeFixtures loads fixtures into Aerospike.
// it is a good idea to use CleanupAerospikeFixtures to clean up after the test.
func LoadAerospikeFixtures(client *aerospike.Client, fixturesData []byte) ([]AerospikeFixture, error) {
	var fixtures []AerospikeFixture
	if err := json.Unmarshal(fixturesData, &fixtures); err != nil {
		return nil, err
	}

	policy := aerospike.NewWritePolicy(0, 0)
	policy.SendKey = true

	for _, fixture := range fixtures {
		key, err := aerospike.NewKey(fixture.Namespace, fixture.Set, fixture.Key)
		if err != nil {
			return nil, err
		}

		binMap := aerospike.BinMap(fixture.Bins)
		if err := client.Put(policy, key, binMap); err != nil {
			return nil, err
		}
	}

	return fixtures, nil
}

// CleanupAerospikeFixtures cleans up fixtures from Aerospike.
func CleanupAerospikeFixtures(client *aerospike.Client, fixtures []AerospikeFixture) error {
	for _, fixture := range fixtures {
		key, err := aerospike.NewKey(fixture.Namespace, fixture.Set, fixture.Key)
		if err != nil {
			return err
		}

		if _, err := client.Delete(nil, key); err != nil {
			return err
		}
	}

	return nil
}
