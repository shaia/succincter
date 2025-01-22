package internal

func CompressToBitVector[T any](input []T, predicate func(T) bool) []uint64 {
	result := make([]uint64, (len(input)+63)/64)
	for i, val := range input {
		if predicate(val) {
			result[i/64] |= 1 << uint(i%64)
		}
	}
	return result
}
