package utils

import "hash/crc32"

func CRC32(field string) uint32 {
	return crc32.Checksum([]byte(field), crc32.MakeTable(crc32.Castagnoli))
}
