package go_rbush

import (
	"testing"
	"fmt"
	"sort"
)

func TestRBush_Load_9_max_entries_by_default(t *testing.T) {
	r := New().Load(someData(9))
	assertEqual(t, r.rootNode.height, 1, "")
	r = New().Load(someData(10))
	assertEqual(t, r.rootNode.height, 2, "")
}

func TestRbush_Search_big_coordinates(t* testing.T) {
	data := [][]float64{
		{-115, 45, -105, 55},
		{105, 45, 115, 55},
		{105, -55, 115, -45},
		{-115, -55, -105, -45},
	}
	tests := []struct{
		intersection BBox
		expected []BBox
	}{
		{
			BBox{MinX: -180, MinY: -90, MaxX: 180, MaxY: 90},
			[]BBox{
				{-115, 45, -105, 55},
				{105, 45, 115, 55},
				{105, -55, 115, -45},
				{-115, -55, -105, -45},
			},
		},
		{
			BBox{MinX: -180, MinY: -90, MaxX: 0, MaxY: 90},
			[]BBox{
			{-115, 45, -105, 55},
			{-115, -55, -105, -45},
			},
		},
		{
		BBox{MinX: 0, MinY: -90, MaxX: 180, MaxY: 90},
		[]BBox{{105, 45, 115, 55},
			{105, -55, 115, -45},
		},},
			{
				BBox{MinX: -180, MinY: 0, MaxX: 180, MaxY: 90},
		[]BBox{{-115, 45, -105, 55},
			{105, 45, 115, 55},
		},},
				{
					BBox{MinX: -180, MinY: -90, MaxX: 180, MaxY: 0},
		[]BBox{{105, -55, 115, -45},
			{-115, -55, -105, -45},
		},},
	}

	for _, d := range (tests) {
		r := New().Load(bboxes(data))
		nodes := r.Search(
			d.intersection,
			// BBox{MinX: -180, MinY: -90, MaxX: 0, MaxY: 90},
		)
		result := make([]BBox, len(nodes))
		for i, n := range(nodes) {
			result[i] = n.BBox
		}
		sorterFactory := func (a []BBox) func (i, j int) bool {
			return func (i, j int) bool {
				if (a[i].MinX != a[j].MinX) {
					return a[i].MinX > a[j].MinX
				}
				return a[i].MinY > a[j].MinY
			}
		}

		sort.Slice(d.expected, sorterFactory(d.expected))
		// For some reason in the original repo it was not necessary to sort the data
		sort.Slice(result, sorterFactory(result))

		for i, _ := range(result) {
			assertEqual(t, result[i], d.expected[i], "")
		}
	}
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

type bboxes [][]float64

func (c bboxes) GetBBoxAt(i int) (x1, y1, x2, y2 float64) {
	return c[i][0], c[i][1], c[i][2], c[i][3]
}

func (c bboxes) Len() int {
	return len(c)
}

func (c bboxes) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c bboxes) Slice(i, j int) Interface {
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
