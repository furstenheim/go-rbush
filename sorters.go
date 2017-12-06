package go_rbush

type pointSorter struct {
	i Interface
}

func (s pointSorter) Less(i, j int) bool {
	x1, y1, _, _ := s.i.GetBBoxAt(i)
	x2, y2, _, _ := s.i.GetBBoxAt(j)
	if x1 < x2 {
		return true
	}
	if x1 == x2 {
		return y1 < y2
	}
	return false
}

func (s pointSorter) Swap(i, j int) {
	s.i.Swap(i, j)
}

func (s pointSorter) Len() int {
	return s.i.Len()
}

type xSorter struct {
	n          *Node
	start, end int
}

func (s xSorter) Less(i, j int) bool {
	x1, _, _, _ := s.n.points.GetBBoxAt(i + s.start)
	x2, _, _, _ := s.n.points.GetBBoxAt(j + s.start)
	return x1 < x2
}

func (s xSorter) Swap(i, j int) {
	s.n.points.Swap(i+s.start, j+s.start)
}

func (s xSorter) Len() int {
	return s.end - s.start
}

type ySorter struct {
	n          *Node
	start, end int
}

func (s ySorter) Less(i, j int) bool {
	_, y1, _, _ := s.n.points.GetBBoxAt(i + s.start)
	_, y2, _, _ := s.n.points.GetBBoxAt(j + s.start)
	return y1 < y2
}

func (s ySorter) Swap(i, j int) {
	s.n.points.Swap(i+s.start, j+s.start)
}

func (s ySorter) Len() int {
	return s.end - s.start
}
