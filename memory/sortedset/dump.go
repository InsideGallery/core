package sortedset

import "encoding/json"

type Dump struct {
	Data    map[string]interface{}
	KeyType string
}

func (s *SortedSet[K, V]) Dump(makeDump func(key K, value V) (string, string, error)) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := map[string][]string{}

	for _, v := range s.dict {
		strKey, strValue, err := makeDump(v.Key(), v.Value())
		if err != nil {
			return "", err
		}

		data[strKey] = append(data[strKey], strValue)
	}

	dump, err := json.Marshal(data)

	return string(dump), err
}

func (s *SortedSet[K, V]) Restore(keyRestore func(key string, values []string) (K, []V, error), dump string) error {
	var data map[string][]string

	err := json.Unmarshal([]byte(dump), &data)
	if err != nil {
		return err
	}

	for key, values := range data {
		rkey, rvalues, err := keyRestore(key, values)
		if err != nil {
			return err
		}

		for _, value := range rvalues {
			s.Upsert(rkey, value)
		}
	}

	return nil
}
