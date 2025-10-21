package simdict

const (
	shingleSize      = 3   // Size of k-grams
	numHashes        = 100 // Number of hash functions for MinHash
	numBands         = 20  // Number of bands for LSH
	rowsPerBand      = 5   // Rows per band (numHashes = numBands * rowsPerBand)
	jaccardThreshold = 0.6 // Similarity threshold to be considered a match
)
