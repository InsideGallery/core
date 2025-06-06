package semver

import (
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/InsideGallery/core/dataconv"
	"github.com/InsideGallery/core/memory/set"
)

type SemVersion struct {
	raw     string
	version [8]uint16
}

func New(version string) (*SemVersion, error) {
	v := &SemVersion{
		raw: version,
	}

	return v, v.build()
}

func (v *SemVersion) build() error {
	version := v.raw
	if version[0] == 'v' {
		version = version[1:]
	}

	prPosition := strings.Index(version, prDelimiter)
	if prPosition == -1 {
		prPosition = len(version)
	}

	buildPosition := strings.Index(version, buildDelimiter)
	if buildPosition == -1 {
		buildPosition = len(version)
	} else {
		v.version[7] = maxValue
	}

	parts, err := v.getCoreParts(version[:prPosition])
	if err != nil {
		return err
	}

	copy(v.version[:3], parts)

	if prPosition+1 < buildPosition {
		parts, err = v.getReleaseParts(version[prPosition+1 : buildPosition])
		if err != nil {
			return err
		}

		copy(v.version[3:7], parts)
	} else {
		v.version[3] = maxValue
		v.version[4] = maxValue
		v.version[5] = maxValue
		v.version[6] = maxValue
	}

	return nil
}

func (v *SemVersion) getCoreParts(version string) ([]uint16, error) {
	var parts []uint16
	var last int

	for i, c := range version {
		if core.Contains(c) {
			val := v.getNumVersion(version, last, i)
			parts = append(parts, val)
			last = i + 1
		}
	}

	val := v.getNumVersion(version, last, len(version))
	parts = append(parts, val)

	return parts, nil
}

func (v *SemVersion) getReleaseParts(version string) ([]uint16, error) {
	var parts []uint16
	var last int

	for i, c := range version {
		if core.Contains(c) {
			num, exist := prOrder[version[last:i]]
			if exist {
				parts = append(parts, num)
				last = i + 1

				continue
			}

			val := v.getNumVersion(version, last, i)
			parts = append(parts, val)
			last = i + 1
		}
	}

	num, exist := prOrder[version[last:]]
	if exist {
		parts = append(parts, num)
		return parts, nil
	}

	val := v.getNumVersion(version, last, len(version))
	parts = append(parts, val)

	return parts, nil
}

func (v *SemVersion) getNumVersion(version string, from, to int) uint16 {
	if from == -1 || to == -1 {
		return 0
	}

	if version[from:to] == "" {
		return 0
	}

	val, err := strconv.ParseUint(version[from:to], base, bitSize)
	if err != nil {
		return 0
	}

	return uint16(val) // nolint gosec
}

func (v *SemVersion) Num() (*big.Int, error) {
	e := dataconv.NewBinaryEncoder()

	for _, val := range v.version {
		err := e.Encode(val)
		if err != nil {
			return nil, err
		}
	}

	num := big.NewInt(0)
	num.SetBytes(e.Bytes())

	return num, nil
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
