package go_rbush

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)
// BenchmarkRBush_Load1Million-4   	       5	1453557818 ns/op	180832276 B/op	 2291974 allocs/op

func BenchmarkRBush_Load1Million(b *testing.B) {
	for i:= 0; i < b.N; i ++ {
		var bigData = getData(1000000, 1)
		tree := NewWithOptions(Options{MAX_ENTRIES: 16}).
			Load(bigData)
		assert.Equal(b, tree.rootNode.height, 5)
	}
}

func randBox(size float64) [4]float64 {
	x := rand.Float64() * (100 - size)
	y := rand.Float64() * (100 - size)
	return [4]float64{x, y,
		x + size*rand.Float64(),
		y + size*rand.Float64()}
}

func getData(N int, size float64) bboxes {
	data := make([][4]float64, N)
	for i := 0; i < N; i++ {
		data[i] = randBox(size)
	}
	return bboxes(data)
}
