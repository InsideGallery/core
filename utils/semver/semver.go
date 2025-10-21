package semver

import (
	"encoding/hex"
	"errors"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/InsideGallery/core/dataconv"
	"github.com/InsideGallery/core/memory/set"
)

type SemVersion struct {
	original string
	version  [8]uint16
	raw      []byte // raw converted version into bytes
}

func New(version string) (*SemVersion, error) {
	v := &SemVersion{
		original: version,
	}

	err := v.build()
	if err != nil {
		return nil, errors.Join(ErrBuildSemver, err)
	}

	b, err := v.Bytes()
	if err != nil {
		return nil, errors.Join(ErrGetRawBytes, err)
	}

	v.raw = b

	return v, nil
}

func (v *SemVersion) build() error {
	version := v.original
	if version[0] == 'v' {
		version = version[1:]
	}

	if version == "" {
		return ErrInvalidVersionString
	}

	prPosition := strings.Index(version, prDelimiter)
	if prPosition == -1 {
		prPosition = len(version)
	}

	buildPosition := strings.Index(version, buildDelimiter)
	if buildPosition == -1 {
		buildPosition = len(version)
	}

	coreEnd := prPosition
	if coreEnd > buildPosition {
		coreEnd = buildPosition
	}

	parts, err := v.getCoreParts(version[:coreEnd])
	if err != nil {
		return err
	}

	copy(v.version[:3], parts)

	if prPosition > 0 && prPosition < len(version) && prPosition+1 >= len(version) {
		return ErrInvalidPreReleaseDelimiter
	}

	if buildPosition > 0 && buildPosition+1 == len(version) {
		return ErrInvalidBuildDelimiter
	}

	if prPosition+1 < buildPosition {
		if version[prPosition+1:buildPosition] == "0" {
			return ErrLeadingZero
		}

		parts, err = v.getReleaseParts(version[prPosition+1 : buildPosition])
		if err != nil {
			return err
		}

		copy(v.version[3:8], parts)
	} else {
		v.version[3] = maxValue
		v.version[4] = 0
		v.version[5] = 0
		v.version[6] = 0
		v.version[7] = 0
	}

	return nil
}

func (v *SemVersion) getCoreParts(version string) ([]uint16, error) {
	var (
		parts []uint16
		last  int
	)

	for i, c := range version {
		if core.Contains(c) {
			val, err := v.getNumVersion(version, last, i)
			if err != nil {
				return nil, err
			}

			parts = append(parts, val)
			last = i + 1
		}
	}

	val, err := v.getNumVersion(version, last, len(version))
	if err != nil {
		return nil, err
	}

	parts = append(parts, val)

	return parts, nil
}

func (v *SemVersion) getReleaseParts(version string) ([]uint16, error) {
	var (
		parts []uint16
		last  int
	)

	for i, c := range version {
		if core.Contains(c) {
			num, exist := prOrder[version[last:i]]
			if exist {
				parts = append(parts, num)
				last = i + 1

				continue
			}

			val, err := v.getNumVersion(version, last, i)
			if err != nil {
				return nil, err
			}

			parts = append(parts, val)
			last = i + 1
		}
	}

	num, exist := prOrder[version[last:]]
	if exist {
		parts = append(parts, num)
		return parts, nil
	}

	val, err := v.getNumVersion(version, last, len(version))
	if err != nil {
		return nil, err
	}

	parts = append(parts, val)

	return parts, nil
}

func (v *SemVersion) getNumVersion(version string, from, to int) (uint16, error) {
	if from == -1 || to == -1 {
		return 0, nil
	}

	if version[from:to] == "" {
		return 0, ErrEmptySegment
	}

	val, err := strconv.ParseUint(version[from:to], base, bitSize)
	if err != nil {
		for _, v := range version[from:to] {
			val += uint64(v)
		}
	}

	if val > math.MaxUint16 {
		return 0, ErrVersionIsOverflow
	}

	return uint16(val), nil // nolint gosec
}

func (v *SemVersion) Bytes() ([]byte, error) {
	e := dataconv.NewBinaryEncoder()

	for _, val := range v.version {
		err := e.Encode(val)
		if err != nil {
			return nil, err
		}
	}

	return e.Bytes(), nil
}

func (v *SemVersion) Hex() string {
	return hex.EncodeToString(v.raw)
}

func (v *SemVersion) Num() *big.Int {
	num := big.NewInt(0)
	num.SetBytes(v.raw)

	return num
}

const (
	maxValue       = math.MaxUint16
	base           = 10
	bitSize        = 64
	prDelimiter    = "-"
	buildDelimiter = "+"
)

var (
	core    = set.NewGenericDataSet('.')
	prOrder = map[string]uint16{
		"alpha": math.MaxUint16 - 4, // nolint mnd
		"beta":  math.MaxUint16 - 3, // nolint mnd
		"pre":   math.MaxUint16 - 2, // nolint mnd
		"rc":    math.MaxUint16 - 1, // nolint mnd
	}
)
