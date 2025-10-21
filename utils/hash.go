package utils

import (
	"hash/crc32"

	"github.com/sirbu/golang-common/hash/crc16"
)

func CRC32(field string) uint32 {
	return crc32.Checksum([]byte(field), crc32.MakeTable(crc32.Castagnoli))
}

func CRC16(field string) uint16 {
	return crc16.Checksum(crc16.XModem, []byte(field))
}
