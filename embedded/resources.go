package embedded

import "embed"

//go:embed resources/*
var fs embed.FS

// GetFS return embed FS
func GetFS() embed.FS {
	return fs
}
