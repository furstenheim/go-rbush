package go_rbush

import (
	"testing"
	"fmt"
)

func TestRBush_Load_9_max_entries_by_default(t *testing.T) {
	r := New().Load(someData(9))
	assertEqual(t, r.rootNode.height, 1, "")
	r = New().Load(someData(10))
	assertEqual(t, r.rootNode.height, 2, "")
}



func someData (n int) coordinates {
	data := make([][2]float64, 0, n)
	for i:= 0; i < n; i++ {
		data = append(data, [2]float64{float64(i), float64(i)})
	}
	return coordinates(data)
}

type coordinates [][2]float64

func (c coordinates) GetBBoxAt(i int) (x1, y1, x2, y2 float64) {
	return c[i][0], c[i][1], c[i][0], c[i][1]
}

func (c coordinates) Len() int {
	return len(c)
}

func (c coordinates) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c coordinates) Slice(i, j int) Interface {
	return c[i:j]
}


func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Errorf(message)
}
