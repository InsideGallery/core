package embedded

import "embed"

//go:embed resources/*
var fs embed.FS

// FS return embed FS
func FS() embed.FS {
	return fs
}

// GetFS return embed FS.
//
// Deprecated: Use FS instead.
func GetFS() embed.FS {
	return fs
}
