package mongodb

import "go.mongodb.org/mongo-driver/bson"

const (
	TwoPrecision = 2
)

// GetBsonD return bson.D object based on values
func GetBsonD(keyValue ...interface{}) bson.D {
	l := len(keyValue)
	if l == 0 || l%TwoPrecision != 0 {
		return bson.D{}
	}

	d := make(bson.D, l/TwoPrecision)

	var k int

	for i := 0; i < len(keyValue); {
		key, val := keyValue[i], keyValue[i+1]
		d[k].Key = key.(string)
		d[k].Value = val

		i += TwoPrecision
		k++
	}

	return d
}
