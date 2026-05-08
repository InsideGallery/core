package embedded

import "embed"

//go:embed resources/*
var fs embed.FS

// FS returns embed FS.
func FS() embed.FS {
	return GetFS()
}

// GetFS returns embed FS.
//
// Deprecated: use FS.
func GetFS() embed.FS {
	return fs
}
