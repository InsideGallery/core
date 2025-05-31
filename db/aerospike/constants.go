package aerospike

const (
	HLLBin = "hll"

	MaxIndexBits          = 13 // between 4 and 16
	MaxAllowedMinhashBits = 8  // between 4 and 51
	// MaxIndexBits+MaxAllowedMinhashBits should be as max 64 bits inclusive
	// sizeof(HLL) = 11 + roundUpToByte(2n_index_bits × (6 + n_minhash_bits))
	// for 12 and 4 we should have 40971 bytes per record
	// for 12 and 8 we should have 57355 bytes per record
	// for 12 and 16 we should have 90123 bytes per record
	// for 12 and 32 we should have 155659 bytes per record
	// for 13 and 4 we should have 81931 bytes per record
	// for 13 and 8 we should have 114688 bytes per record
)
