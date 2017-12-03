// This package implements a geometric rbush index based on https://github.com/mourner/rbush
package go_rbush

// Interface abstract the required properties for an slice of points
import (
	"sort"
	"log"
	"math"
)


const (
	MAX_ENTRIES = 9
	MIN_ENTRIES = 4
	NUMBER_OF_SORTERS
)
type Interface interface {
	Take(i int) Point         // Retrieve point at position i
	Len() int                 // Number of elements
	Swap(i, j int)            // Swap elements with indexes i and j
	Slice(i, j int) Interface //Slice the interface between two indices

}

// A point basically returns coordinates
type Point interface {
	GetCoordinates() (float64, float64)
}

// Create an RBush index from an array of points
func New (points Interface) RBush {
	sort.Sort(pointSorter{i: points})
	return NewFromSortedArray(points)
}

// Create an RBush index from an array of points which is already in lexicographical order
func NewFromSortedArray (points Interface) RBush {
	r := RBush{points: points, rootNode:nil}
}

type RBush struct {
	rootNode *Node
	points Interface
}

type BBox struct {

}

type Node struct {
	children []*Node
	start, end, height int
	isLeaf bool
	points Interface
}

func (r RBush) Search (b BBox) {

}

func (r RBush) Collides (b BBox) {

}

func (r *RBush) Load (points Interface) {
	node := r.build(points, 0, points.Len(), 0)
	if (r.rootNode == nil){
		r.rootNode = node
	} else {
		// TODO
	}
}
// points is assumed to be ordered
func (r *RBush) build (points Interface, start, end, height int) Node {
	ch := make(chan *Node)
	readCh := make(chan *Node)
	exitCh := make(chan int, NUMBER_OF_SORTERS)
	for i:= 0; i < NUMBER_OF_SORTERS; i++ {
		go func (ch, readCh chan *Node, exitCh chan int) {
			for true {
				select {
				case n := <- ch:
					N := n.end - n.start + 1
					// target number of root entries to maximize storage utilization
					var M float64
					if (N <= MAX_ENTRIES) { // Leaf node
						// TODO calcbox, maybe associate slice
						n.isLeaf = true
						readCh <- n
						continue
					}
					// sort on x, then split in equal size buckets and sort in y
					// root node is assumed sorted so there is no need to sort
					// first node inserted
					if (n.height == -1) {
						// This is the target height
						n.height = math.Ceil(math.Log(N) / math.Log(MAX_ENTRIES))
					} else {
						sortX := xSorter{n: n, start: n.start, end: n.end}
						sort.Sort(sortX)
					}
					M = math.Ceil(float64(N) / float64(math.Pow(MAX_ENTRIES, n.height - 1)))

					N2 := math.Ceil(N / M)
					N1 := N2 * math.Ceil(math.Sqrt(M))
					for i := n.start; i <= n.end; i += N1 {
						right2 := math.Min(i + N - 1, n.end)
						sortY := ySorter{n: n, start: i, end: right2}
						sort.Sort(sortY)
						for j := i; j <= right2; j += N2 {
							right3 := math.Min(j + N2 - 1, right2)
							child := Node{
								start: j,
								end: right3,
								points: n.points,
								height: n.height - 1,
							}
							n.children = append(n.children, child)
							readCh <- child
						}
					}

				case <-exitCh:
					return
				}
			}

		}(ch, readCh, exitCh)
	}

	remainingNodes := 1

	rootNode := Node{start: start, end: end, height: -1, points: points}
	ch <- &rootNode

	// TODO insert
	for remainingNodes > 0 {
		select {
		// nodes here are received already ordered
		case n:= <- readCh:
			remainingNodes -= 1
			for _, childNode := range(n.children) {
				remainingNodes += 1
				ch <- childNode
			}
		}
	}
	for i:= 0; i < NUMBER_OF_SORTERS; i++ {
		exitCh <- 1
	}
	return rootNode
}

func (r RBush) Insert () {

}

func (r RBush) Clear () {

}

func (r RBush) Remove () {

}

func (r RBush) ToBBox () {

}

