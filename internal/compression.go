package internal

func CompressToBitvector(input []bool) []uint64 {
	var bitvector []uint64
	current := uint64(0)
	for i, bit := range input {
		if bit {
			current |= 1 << (i % 64)
		}
		if (i+1)%64 == 0 {
			bitvector = append(bitvector, current)
			current = 0
		}
	}
	if len(input)%64 != 0 {
		bitvector = append(bitvector, current)
	}
	return bitvector
}
