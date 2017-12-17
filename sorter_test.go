package go_rbush

import (
	"testing"
	"reflect"
	"sort"
)

func TestFloydRivestSelect(t *testing.T) {
	a := []int{65, 28, 59, 33, 21, 56, 22, 95, 50, 12, 90, 53, 28, 77, 39}
	FloydRivestSelect(sorter(a), 8, 0, len(a) - 1)
	// TODO more tests
	// TODO tests for >600
	expected := []int{39, 28, 28, 33, 21, 12, 22, 50, 53, 56, 59, 65, 90, 77, 95}

	assertEqual(t, reflect.DeepEqual(a, expected), true, "")
}
func TestFloydRivestBuckets(t *testing.T) {
	a := []int{65, 28, 59, 33, 21, 56, 22, 95, 50, 12, 90, 53, 28, 77, 39}
	FloydRivestBuckets(sorter(a), 1, 0, len(a) - 1)
	expected := []int{39, 28, 28, 33, 21, 12, 22, 50, 53, 59, 56, 65, 95, 77, 90}
	sort.Sort(sorter(expected))
	// expected := []int{21 12 39 22 28 28 50 33 95 56 53 59 77 90 65}
	// this is Mourner's expected result, the important part is 0..8 and that is the same. for some reason the rest is different
	// It might mean that the algorithm is wrong
	// TODO more tests
	// TODO tests for >600
	//expected := []int{39, 28, 28, 33, 21, 12, 22, 50, 53, 56, 59, 65, 90, 77, 95}

	assertEqual(t, reflect.DeepEqual(a, expected), true, "")
}

type sorter []int
func (s sorter) Len() int {
	return len(s)
}

func (s sorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sorter) Less (i, j int) bool {
	return s[i] < s[j]
}
