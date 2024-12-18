package core

// isSimilar compares two hashes with a similarity threshold
func IsSimilar(hash1, hash2 []byte, errorRate float64) bool {
	distance := hammingDistance(hash1, hash2)
	return float64(distance)/float64(len(hash1)*8) <= errorRate
}

// hammingDistance computes the Hamming distance between two hashes
func hammingDistance(hash1, hash2 []byte) int {
	var distance int
	for i := 0; i < len(hash1); i++ {
		distance += popCount(hash1[i] ^ hash2[i])
	}
	return distance
}

// popCount counts bits set to 1
func popCount(x byte) int {
	count := 0
	for x > 0 {
		count++
		x &= x - 1
	}
	return count
}
