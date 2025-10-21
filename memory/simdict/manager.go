package simdict

type LSHManager struct {
	lshIndex      *LSHIndex
	docSignatures map[string][]uint32 // Stores original signatures {docID -> signature}
	docBuckets    map[string]string   // Maps each docID to its bucketID {docID -> bucketID}
	bucketMembers map[string][]string // Maps a bucketID to all its members {bucketID -> [docIDs]}
}

func NewLSHManager() *LSHManager {
	return &LSHManager{
		lshIndex:      NewLSHIndex(numBands, rowsPerBand),
		docSignatures: make(map[string][]uint32),
		docBuckets:    make(map[string]string),
		bucketMembers: make(map[string][]string),
	}
}

func (m *LSHManager) ProcessAndAssign(docID string) string {
	// 1. Generate Signature
	shingles := shingle(docID, shingleSize)
	signature := createSignature(shingles, numHashes)

	// 2. Query for Candidates
	candidates := m.lshIndex.Query(signature)

	var bucketID string

	// 3. Verify Similarity with Candidates
	for _, candidateID := range candidates {
		// Retrieve the candidate's signature
		candidateSig, ok := m.docSignatures[candidateID]
		if !ok {
			continue // Should not happen in a consistent state
		}

		// Calculate precise similarity
		similarity := jaccardFromSignatures(signature, candidateSig)

		if similarity >= jaccardThreshold {
			// Found a match! Assign to this candidate's bucket.
			bucketID = m.docBuckets[candidateID]
			break
		}
	}

	// 4. Assign to a bucket
	if bucketID == "" {
		// No similar items found, this document starts a new bucket.
		bucketID = docID // The first item in a bucket is its ID.
	}

	// 5. Update all data structures
	// Check if already processed to avoid re-adding
	if _, exists := m.docSignatures[docID]; !exists {
		m.lshIndex.Add(docID, signature)
		m.docSignatures[docID] = signature
		m.docBuckets[docID] = bucketID
		m.bucketMembers[bucketID] = append(m.bucketMembers[bucketID], docID)
	}

	return bucketID
}
