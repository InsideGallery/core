package semver

import "errors"

var (
	ErrBuildSemver = errors.New("err execute build")
	ErrGetRawBytes = errors.New("err get raw bytes")
)
