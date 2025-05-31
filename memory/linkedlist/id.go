package linkedlist

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (l *List[V]) NextID() string {
	return primitive.NewObjectID().Hex()
}
