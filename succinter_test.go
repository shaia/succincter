package succincter

import (
	"testing"
)

func TestSuccincterRank(t *testing.T) {
	input := []bool{true, false, true, true, false, true, false, false, true}
	succincter := NewSuccincter(input)

	result := succincter.Rank(5)
	// Should output the number of 1's up to position 5
	if result != 3 {
		t.Error("Rank(5) should be 3")
	}

}

func TestSuccincterSelect(t *testing.T) {
	input := []bool{true, false, true, true, false, true, false, false, true}
	succincter := NewSuccincter(input)

	result := succincter.Select(3)
	// Should output the position of the 3rd 1
	if result != 3 {
		t.Error("Rank(5) should be 5")
	}
}
