package semver

import "errors"

var (
	ErrBuildSemver       = errors.New("err execute build")
	ErrGetRawBytes       = errors.New("err get raw bytes")
	ErrVersionIsOverflow = errors.New("err version is overflow")

	ErrInvalidVersionString       = errors.New("invalid semantic version string")
	ErrInvalidPreReleaseDelimiter = errors.New("invalid pre-release delimiter")
	ErrInvalidBuildDelimiter      = errors.New("invalid build delimiter")
	ErrInvalidCharacter           = errors.New("invalid character in version string")
	ErrLeadingZero                = errors.New("version segment has leading zero")
	ErrEmptySegment               = errors.New("version contains an empty segment")
)
