// This package implements a geometric rbush index based on https://github.com/mourner/rbush
package go_rbush

// Interface abstract the required properties for an slice of points
import (
	"log"
	"math"
	"runtime"
)

const (
	MIN_ENTRIES         = 4
	MAX_HEIGHT_TO_SPLIT = 3 // When creating the index we'll split the task into a new goroutine until we reach this height
)

type Interface interface {
	GetBBoxAt(i int) (x1, y1, x2, y2 float64) // Retrieve point at position i
	Len() int                                 // Number of elements
	Swap(i, j int)                            // Swap elements with indexes i and j
	Slice(i, j int) Interface                 //Slice the interface between two indices
}

type Options struct {
	MAX_ENTRIES int
}

// Create an RBush index from an array of points
func New() *RBush {
	defaultOptions := Options{
		MAX_ENTRIES: 9,
	}
	return NewWithOptions(defaultOptions)
}

func NewWithOptions(options Options) *RBush {
	r := &RBush{
		options: options,
	}
	r.initRootNode()
	return r
}

type RBush struct {
	options  Options
	rootNode *Node
}

type Node struct {
	children   []*Node
	height     int
	isLeaf     bool
	points     Interface
	parentNode *Node
	BBox       BBox
}

func (r *RBush) Search(b BBox) []*Node {
	// TODO remove
	_ = runtime.GOOS
	node := r.rootNode
	result := make([]*Node, 0)
	if !node.BBox.intersects(b) {
		return result
	}
	nodesToSearch := make([]*Node, 0, 1)
	nodesToSearch = append(nodesToSearch, node)
	for len(nodesToSearch) != 0 {
		// pop first item
		node, nodesToSearch = nodesToSearch[0], nodesToSearch[1:]
		for _, c := range node.children {
			if b.intersects(c.BBox) {
				if node.isLeaf {
					// child is basically a point
					result = append(result, c)
				} else if b.contains(c.BBox) {
					result = append(result, c.flattenDownwards()...)
				} else {
					nodesToSearch = append(nodesToSearch, c)
				}
			}
		}
	}
	return result
}

func (r *RBush) Collides(b BBox) bool {
	node := r.rootNode
	if !node.BBox.intersects(b) {
		return false
	}
	nodesToSearch := make([]*Node, 0, 10)
	nodesToSearch = append(nodesToSearch, node)
	for len(nodesToSearch) != 0 {
		// pop first item
		node, nodesToSearch = nodesToSearch[0], nodesToSearch[1:]
		for _, c := range node.children {
			// TODO leaf
			if c.BBox.intersects(b) {
				if node.isLeaf || b.contains(c.BBox) {
					return true
				}
				nodesToSearch = append(nodesToSearch, c)
			}
		}
	}
	return false
}

// Returns all end points inside node
func (n *Node) flattenDownwards() []*Node {
	var node *Node
	// runtime.Breakpoint()
	result := make([]*Node, 0, len(n.children))
	nodesToSearch := []*Node{n}
	for len(nodesToSearch) != 0 {
		node, nodesToSearch = nodesToSearch[0], nodesToSearch[1:]
		if node.isLeaf {
			result = append(result, node.children...)
		} else {
			nodesToSearch = append(nodesToSearch, node.children...)
		}
	}
	return result
}

func (r *RBush) Load(points Interface) *RBush {
	return r.load(points, false)
}

func (r *RBush) LoadSortedArray(points Interface) *RBush {
	return r.load(points, true)
}

func (r *RBush) load (points Interface, isSorted bool) *RBush {
	if points.Len() == 0 {
		return r
	}

	if points.Len() < MIN_ENTRIES {
		for i := 0; i < points.Len(); i++ {
			r.InsertElement(points.Slice(i, i+1))
		}
		return r
	}
	// TODO points.Len < MIN_ENTRIEs
	node := r.build(points, isSorted)
	if len(r.rootNode.children) == 0 {
		r.rootNode = node
	} else if r.rootNode.height == node.height {
		r.splitRoot(node)
	} else {
		if r.rootNode.height < node.height {
			// swap nodes and insert smaller one
			tmpNode := r.rootNode
			r.rootNode = node
			node = tmpNode
		}
		// insert small tree into big tree
		r.insertNode(node)
	}

	return r
}

// points is assumed to be ordered
func (r *RBush) build(points Interface, isSorted bool) *Node {

	confirmCh := make(chan int, 1)

	rootNode := &Node{
		height: int(math.Ceil(math.Log(float64(points.Len())) / math.Log(float64(r.options.MAX_ENTRIES)))),
		points: points}
	remainingNodes := 1

	go r.buildNodeDownwards(rootNode, confirmCh, true, isSorted)
	for remainingNodes > 0 {
		i := <-confirmCh
		remainingNodes += i
	}
	close(confirmCh)
	rootNode.computeBBoxDownwards()
	return rootNode
}

func (r *RBush) buildNodeDownwards(n *Node, confirmCh chan int, isCalledAsync, isSorted bool) {
	if isCalledAsync {
		defer func() {
			confirmCh <- -1
		}()
	}

	N := n.points.Len()
	// target number of root entries to maximize storage utilization
	var M float64
	if N <= r.options.MAX_ENTRIES { // Leaf node
		n.setLeafNode(n.points)
		return
	}

	M = math.Ceil(float64(N) / float64(math.Pow(float64(r.options.MAX_ENTRIES), float64(n.height-1))))

	N2 := int(math.Ceil(float64(N) / M))
	N1 := N2 * int(math.Ceil(math.Sqrt(M)))

	// parent node might already be sorted. In that case we avoid double computation
	if (n.parentNode != nil || !isSorted) {
		sortX := xSorter{n: n, start: 0, end: n.points.Len(), bucketSize:  N1}
		sortX.Sort()
	}
	// runtime.Breakpoint()
	for i := 0; i < n.points.Len(); i += N1 {
		right2 := minInt(i+N1, n.points.Len())
		sortY := ySorter{n: n, start: i, end: right2, bucketSize: N2}
		sortY.Sort()
		for j := i; j < right2; j += N2 {
			right3 := minInt(j+N2, right2)
			child := Node{
				points:     n.points.Slice(j, right3),
				height:     n.height - 1,
				parentNode: n,
			}
			n.children = append(n.children, &child)
			// remove reference to interface, we only need it for points

		}
	}
	n.points = nil
	// compute children
	for _, c := range n.children {
		// Only launch a goroutine for big height. we don't want a goroutine to sort 4 points
		if n.height > MAX_HEIGHT_TO_SPLIT {
			confirmCh <- 1
			go r.buildNodeDownwards(c, confirmCh, true, false)
		} else {
			r.buildNodeDownwards(c, confirmCh, false, false)
		}
	}
}


func (r *RBush) InsertElement(p Interface) {
	x1, y1, x2, y2 := p.GetBBoxAt(0)
	node := Node{
		points: p,
		BBox: BBox{
			MinX: x1,
			MaxX: x2,
			MinY: y1,
			MaxY: y2,
		},
	}
	// TODO make sure this actually works
	r.insertNode(&node)
}

func (r *RBush) insertNode(n *Node) {
	// insert small tree into big tree
	chosenNode := r.chooseSubtree(n)
	// TODO probably do something in the case : n.isLeaf, chosenNode.isLeaf
	n.parentNode = chosenNode
	chosenNode.children = append(chosenNode.children, n)
	chosenNode.BBox = chosenNode.BBox.extend(n.BBox)

	// split on node overflow, propagate upwards
	for iterNode := chosenNode; iterNode != nil; iterNode = iterNode.parentNode {
		if len(iterNode.children) > r.options.MAX_ENTRIES {
			r.split(iterNode)
		} else {
			iterNode.BBox = iterNode.BBox.extend(n.BBox)
		}
	}

}

func (r *RBush) splitRoot(n *Node) {
	newRoot := Node{
		height: r.rootNode.height + 1,
		children: []*Node{
			r.rootNode,
			n,
		},
	}
	r.rootNode.parentNode = &newRoot
	n.parentNode = &newRoot
	r.rootNode = &newRoot
}

func (n *Node) setLeafNode(p Interface) {
	// Here we follow original rbush implementation.
	// TODO try to store elements children as points instead of nodes
	// It seems a bit inefficient to have one child for each point, but otherwise the complexity of the code blows up
	children := make([]*Node, p.Len())
	n.children = children
	n.points = nil
	n.height = 1
	n.isLeaf = true

	for i := 0; i < p.Len(); i++ {
		x1, y1, x2, y2 := p.GetBBoxAt(i)
		children[i] = &Node{
			points: p.Slice(i, i+1),
			BBox: BBox{
				MinX: x1,
				MaxX: x2,
				MinY: y1,
				MaxY: y2,
			},
			parentNode: n,
		}
	}
}

// split node into two, update bboxes
func (r *RBush) split(n *Node) {
	n.chooseSplitAxis()
	i := n.chooseSplitIndex()
	newNode := Node{
		children:   n.children[i:len(n.children)],
		height:     n.height,
		parentNode: n.parentNode,
		isLeaf:     n.isLeaf,
	}
	n.children = n.children[0:i]
	for _, c := range newNode.children {
		c.parentNode = &newNode
	}
	n.BBox = n.partialBBox(0, len(n.children))
	newNode.BBox = newNode.partialBBox(0, len(newNode.children))
	// not root
	if n.parentNode != nil {
		n.parentNode.children = append(n.parentNode.children, &newNode)
	} else {
		r.splitRoot(&newNode)
	}

}

// sorts children by best axis for split
func (n *Node) chooseSplitAxis() {
	// TODO do properly
}

// find best index to split
func (n *Node) chooseSplitIndex() int {
	// TODO do properly
	return len(n.children) / 2
}

// find optimal node searching for the node that grows less in area.
func (r *RBush) chooseSubtree(n *Node) *Node {
	// -1 because we want the node to be at the same level
	// n.height same as rootNode.height is not considered here since we would have called split root
	requiredDepth := r.rootNode.height - n.height - 1
	if requiredDepth < 0 {
		// Most definitely an error in the implementation
		log.Fatal("We are inserting a big tree into a smaller tree.")
	}
	depth := 0
	chosenNode := r.rootNode
	for true {
		// We always insert small tree into big tree so it cannot happen that we insert a non point into a leaf
		if chosenNode.isLeaf || depth == requiredDepth {
			break
		}
		minArea := math.Inf(+1)
		minEnlargement := math.Inf(+1)
		var targetNode *Node
		for _, child := range chosenNode.children {
			area := child.BBox.area()
			enlargement := n.BBox.enlargedArea(child.BBox) - area

			// find entry with minimum enlargment
			if enlargement < minEnlargement {
				minEnlargement = enlargement
				if area < minArea {
					minArea = area
				}
				targetNode = child
			} else if enlargement == minEnlargement {
				if area < minArea {
					minArea = area
					targetNode = child
				}
			}
		}
		if targetNode != nil {
			chosenNode = targetNode
		} else {
			// in case we cannot choose among all children (for example if area is infinity then we chose first child)
			chosenNode = chosenNode.children[0]
		}
		depth++
	}
	return chosenNode

}

// Compute bbox of all tree all the way to the bottom
func (n *Node) computeBBoxDownwards() BBox {

	var bbox BBox
	if n.isLeaf {
		bbox = BBox{
			MinX: math.Inf(+1),
			MaxX: math.Inf(-1),
			MinY: math.Inf(+1),
			MaxY: math.Inf(-1),
		}
		// This bounded boxes are computed when creating the nodes, they only contain one point so there is no doubt
		for i := 0; i < len(n.children); i++ {
			bbox = bbox.extend(n.children[i].BBox)
		}
	} else {
		bbox = n.children[0].computeBBoxDownwards()

		for i := 1; i < len(n.children); i++ {
			bbox = bbox.extend(n.children[i].computeBBoxDownwards())
		}
	}
	n.BBox = bbox
	return bbox
}

// compute bbox of part of the childre
func (n *Node) partialBBox(start, end int) BBox {
	bbox := n.children[start].BBox
	for i := start + 1; i < end; i++ {
		bbox = bbox.extend(n.children[i].BBox)
	}
	return bbox
}

func (r *RBush) Clear() {

}

func (r *RBush) initRootNode() {
	r.rootNode = &Node{
		children: []*Node{},
		BBox: BBox{
			MinX: math.Inf(1),
			MaxX: math.Inf(-1),
			MinY: math.Inf(1),
			MaxY: math.Inf(-1),
		},
		isLeaf: true,
		height: 1,
	}
}


func (r *RBush) Remove(p ToBeRemoved) (*RBush){
	x1, y1, x2, y2 := p.GetBBox()
	bbox := BBox{x1, y1, x2, y2}
	node := r.rootNode
	nodesToSearch := make([]*Node, 0, 1)
	indexes := make([]int, 0, r.rootNode.height) // keep track of indices to clean
	currentIndex := 0 // index at current level
	isGoingUp := false
	nodesToSearch = append(nodesToSearch, node)
	for node != nil {
		if (node.isLeaf) {
			index := node.findIndexToRemove(p)
			if (index != -1) {
				indexes = append(indexes, currentIndex, index)
				// remove first index because we always have 0 for root node
				node.children[index].removeUpwards(indexes[1:])
				return r
			}
		}

		if (!isGoingUp && !node.isLeaf && node.BBox.contains(bbox)) { // go down
			indexes = append(indexes, currentIndex)
			currentIndex = 0
			node = node.children[0] // TODO check size 0
		} else if node.parentNode != nil && currentIndex < len(node.parentNode.children) - 1 { // go sideways
			currentIndex++
			node = node.parentNode.children[currentIndex]
			isGoingUp = false
		} else {
			// we found nothing. Go up
			node = node.parentNode
			isGoingUp = true
			indexes, currentIndex = indexes[0: len(indexes) -1], indexes[len(indexes) -1]
		}
	}
	return r
}

func (n *Node) findIndexToRemove (p ToBeRemoved) (int) {
	index := -1
	x1, y1, x2, y2 := p.GetBBox()
	bbox := BBox{x1, y1, x2, y2}
	// Maybe we can do something fancier since points might be ordered
	for i, c := range(n.children) {
		if bbox.equals(c.BBox) && p.IsContained(c.points) {
			return i
		}
	}
	return index
}

// Remove node from parent and condense if necessary. Update bboxes
// indexes represent indexes in parent
func (n *Node) removeUpwards(indexes []int) {
	nodeToRemove := n
	parent := n.parentNode
	for i := len(indexes) - 1; i >= 0; i-- {
		index := indexes[i]
		if nodeToRemove != nil {
			var removedNode *Node
			removedNode = parent.children[index]
			parent.children = append(parent.children[0: index], parent.children[index + 1: len(parent.children)]...)
			if (nodeToRemove != removedNode) {
				log.Fatal("Removed incorrect node")
			}
			if len(parent.children) == 0 {
				nodeToRemove = parent
			} else {
				nodeToRemove = nil
			}
		}
		parent.BBox = parent.partialBBox(0, len(parent.children))
		parent = parent.parentNode
	}
}

type ToBeRemoved interface {
	GetBBox () (x1, y1, x2, y2 float64)
	IsContained (points Interface) bool
}

func (r RBush) ToBBox() {

}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
