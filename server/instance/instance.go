package instance

import (
	"log"

	"github.com/InsideGallery/core/utils"
)

var (
	id      = utils.GetUniqueID()
	shortID string
)

func init() {
	sid, err := utils.GetShortID()
	if err != nil {
		log.Fatalf("Error getting short id: %v", err)
	}

	shortID = string(sid)
}

// GetInstanceID return current instance id
func GetInstanceID() string {
	return id
}

// GetShortInstanceID return current short instance id
func GetShortInstanceID() string {
	return shortID
}
