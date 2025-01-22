package internal

type SimpleArray struct {
	data []bool
}

func NewSimpleArray(input []bool) *SimpleArray {
	return &SimpleArray{
		data: append([]bool{}, input...),
	}
}

func (sa *SimpleArray) Rank(pos int) int {
	if pos <= 0 {
		return 0
	}
	if pos > len(sa.data) {
		pos = len(sa.data)
	}
	count := 0
	for i := 0; i < pos; i++ {
		if sa.data[i] {
			count++
		}
	}
	return count
}

func (sa *SimpleArray) Select(rank int) int {
	if rank <= 0 {
		return -1
	}
	count := 0
	for i := range sa.data {
		if sa.data[i] {
			count++
			if count == rank {
				return i
			}
		}
	}
	return -1
}
