package instance

import (
	"log/slog"
	"sync"

	"github.com/InsideGallery/core/utils"
)

var (
	id      = utils.GetUniqueID()
	shortID string
	once    sync.Once
)

func initShortID() {
	once.Do(func() {
		sid, err := utils.GetShortID()
		if err != nil {
			slog.Default().Error("get short instance id", "err", err)
			return
		}

		shortID = string(sid)
	})
}

// GetInstanceID return current instance id
func GetInstanceID() string {
	return id
}

// GetShortInstanceID return current short instance id
func GetShortInstanceID() string {
	initShortID()

	return shortID
}
