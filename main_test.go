package go_rbush

import (
	"fmt"
	"math"
	"sort"
	"testing"
)

func TestRBush_Load_9_max_entries_by_default(t *testing.T) {
	r := New().Load(someData(9))
	assertEqual(t, r.rootNode.height, 1, "")
	r = New().Load(someData(10))
	assertEqual(t, r.rootNode.height, 2, "")
}

func TestRBush_Search_big_coordinates(t *testing.T) {
	data := [][4]float64{
		{-115, 45, -105, 55},
		{105, 45, 115, 55},
		{105, -55, 115, -45},
		{-115, -55, -105, -45},
	}
	tests := []struct {
		intersection BBox
		expected     []BBox
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
			}},
		{
			BBox{MinX: -180, MinY: 0, MaxX: 180, MaxY: 90},
			[]BBox{{-115, 45, -105, 55},
				{105, 45, 115, 55},
			}},
		{
			BBox{MinX: -180, MinY: -90, MaxX: 180, MaxY: 0},
			[]BBox{{105, -55, 115, -45},
				{-115, -55, -105, -45},
			}},
	}

	for _, d := range tests {
		r := New().Load(bboxes(data))
		nodes := r.Search(
			d.intersection,
			// BBox{MinX: -180, MinY: -90, MaxX: 0, MaxY: 90},
		)
		result := make([]BBox, len(nodes))
		for i, n := range nodes {
			result[i] = n.BBox
		}
		sorterFactory := func(a []BBox) func(i, j int) bool {
			return func(i, j int) bool {
				if a[i].MinX != a[j].MinX {
					return a[i].MinX > a[j].MinX
				}
				return a[i].MinY > a[j].MinY
			}
		}

		sort.Slice(d.expected, sorterFactory(d.expected))
		// For some reason in the original repo it was not necessary to sort the data
		sort.Slice(result, sorterFactory(result))

		for i, _ := range result {
			assertEqual(t, result[i], d.expected[i], "")
		}
	}
}

func TestRBush_Load(t *testing.T) {
	data := getDataExample()
	originalData := getDataExample()
	childNodes := New().Load(data).rootNode.flattenDownwards()
	recoveredPoints := make([][4]float64, 0, len(childNodes))
	for _, n := range childNodes {
		b := [][4]float64(n.points.(bboxes))
		recoveredPoints = append(recoveredPoints, b...)
	}

	sort.Sort(bboxes(recoveredPoints))
	sort.Sort(bboxes(originalData))
	assertEqual(t, len(recoveredPoints), len(originalData), fmt.Sprintf("We should get the same amout of points, %v %v", len(recoveredPoints), len(originalData)))
	if len(recoveredPoints) != len(originalData) {
		fmt.Println(recoveredPoints)
		fmt.Println(originalData)
		return
	}
	for i, _ := range recoveredPoints {
		assertEqual(t, originalData[i], recoveredPoints[i], "")
	}
}

func TestRBush_Loading_little_data(t *testing.T) {
	data := getDataExample()
	// It is important to use another slice so they don't share the same (we are modifying the underlying array)
	data2 := getDataExample()
	// In the original test they compare the trees themselves. Maybe we could do the same walking the tree
	tree1 := New().
		Load(data).
		Load(data2[0:3])
	tree2 := New().
		Load(data).
		Load(data2[0:1]).
		Load(data2[1:2]).
		Load(data2[2:3])
	recoveredPoints1 := getTreePointsAsCoordinates(tree1.rootNode)
	recoveredPoints2 := getTreePointsAsCoordinates(tree2.rootNode)

	assertEqual(t, len(recoveredPoints1), len(recoveredPoints2), fmt.Sprintf("We should get the same amout of points, %v %v", len(recoveredPoints1), len(recoveredPoints2)))
	if len(recoveredPoints1) != len(recoveredPoints2) {
		return
	}
	for i, _ := range recoveredPoints1 {
		assertEqual(t, recoveredPoints1[i], recoveredPoints2[i], "")
	}
}

func TestRBush_LoadNothing(t *testing.T) {
	tree1 := New().Load(make(coordinates, 0))
	assertEqual(t, len(tree1.rootNode.children), 0, "Root has not children on init mode")
}

func TestRBush_LoadEmptyData(t *testing.T) {
	data := getEmptyDataExample()
	tree := NewWithOptions(Options{
		MAX_ENTRIES: 4,
	}).Load(data)
	assertEqual(t, tree.rootNode.height, 2, "")
	recoveredPoints := getTreePointsAsCoordinates(tree.rootNode)
	assertEqual(t, len(recoveredPoints), len(data),
		fmt.Sprintf("We should get the same amout of points, %v %v", len(data), len(recoveredPoints)))
	if len(recoveredPoints) != len(data) {
		return
	}
	for i, _ := range recoveredPoints {
		assertEqual(t, recoveredPoints[i], data[i], "")
	}
}

func TestRBush_LoadEmptyDataElementByElement(t *testing.T) {
	data := getEmptyDataExample()
	tree := NewWithOptions(Options{
		MAX_ENTRIES: 4,
	})
	for i, _ := range data {
		tree.InsertElement(data[i : i+1])
	}

	assertEqual(t, tree.rootNode.height, 2, "")
	recoveredPoints := getTreePointsAsCoordinates(tree.rootNode)
	assertEqual(t, len(recoveredPoints), len(data),
		fmt.Sprintf("We should get the same amout of points, %v %v", len(data), len(recoveredPoints)))
	if len(recoveredPoints) != len(data) {
		fmt.Println(data)
		fmt.Println(recoveredPoints)
		return
	}
	for i, _ := range recoveredPoints {
		assertEqual(t, recoveredPoints[i], data[i], "")
	}
}

func getTreePointsAsCoordinates(n *Node) [][4]float64 {
	childNodes := n.flattenDownwards()
	recoveredPoints := make([][4]float64, 0, len(childNodes))
	for _, c := range childNodes {
		b := [][4]float64(c.points.(bboxes))
		recoveredPoints = append(recoveredPoints, b...)
	}
	sort.Sort(bboxes(recoveredPoints))
	return recoveredPoints

}
func someData(n int) coordinates {
	data := make([][2]float64, 0, n)
	for i := 0; i < n; i++ {
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

type bboxes [][4]float64

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

// to sort them for tests
func (c bboxes) Less(i, j int) bool {
	if c[i][0] != c[j][0] {
		return c[i][0] < c[j][0]
	}
	return c[i][1] < c[j][1]
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

func getDataExample() bboxes {
	return bboxes{{0, 0, 0, 0}, {10, 10, 10, 10}, {20, 20, 20, 20}, {25, 0, 25, 0}, {35, 10, 35, 10}, {45, 20, 45, 20}, {0, 25, 0, 25}, {10, 35, 10, 35},
		{20, 45, 20, 45}, {25, 25, 25, 25}, {35, 35, 35, 35}, {45, 45, 45, 45}, {50, 0, 50, 0}, {60, 10, 60, 10}, {70, 20, 70, 20}, {75, 0, 75, 0},
		{85, 10, 85, 10}, {95, 20, 95, 20}, {50, 25, 50, 25}, {60, 35, 60, 35}, {70, 45, 70, 45}, {75, 25, 75, 25}, {85, 35, 85, 35}, {95, 45, 95, 45},
		{0, 50, 0, 50}, {10, 60, 10, 60}, {20, 70, 20, 70}, {25, 50, 25, 50}, {35, 60, 35, 60}, {45, 70, 45, 70}, {0, 75, 0, 75}, {10, 85, 10, 85},
		{20, 95, 20, 95}, {25, 75, 25, 75}, {35, 85, 35, 85}, {45, 95, 45, 95}, {50, 50, 50, 50}, {60, 60, 60, 60}, {70, 70, 70, 70}, {75, 50, 75, 50},
		{85, 60, 85, 60}, {95, 70, 95, 70}, {50, 75, 50, 75}, {60, 85, 60, 85}, {70, 95, 70, 95}, {75, 75, 75, 75}, {85, 85, 85, 85}, {95, 95, 95, 95}}
}

func getEmptyDataExample() bboxes {
	return bboxes{{math.Inf(-1), math.Inf(-1), math.Inf(+1), math.Inf(+1)}, {math.Inf(-1), math.Inf(-1), math.Inf(+1), math.Inf(+1)},
		{math.Inf(-1), math.Inf(-1), math.Inf(+1), math.Inf(+1)}, {math.Inf(-1), math.Inf(-1), math.Inf(+1), math.Inf(+1)},
		{math.Inf(-1), math.Inf(-1), math.Inf(+1), math.Inf(+1)}, {math.Inf(-1), math.Inf(-1), math.Inf(+1), math.Inf(+1)}}
}
