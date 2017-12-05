// This package implements a geometric rbush index based on https://github.com/mourner/rbush
package go_rbush

// Interface abstract the required properties for an slice of points
import (
	"log"
	"math"
	"sort"
)

const (
	MAX_ENTRIES       = 9
	MIN_ENTRIES       = 4
	NUMBER_OF_SORTERS = 4
)

type Interface interface {
	GetCoordinatesAt(i int) (x, y float64)                        // Retrieve point at position i
	Len() int                                // Number of elements
	Swap(i, j int)                           // Swap elements with indexes i and j
	Slice(i, j int) Interface                //Slice the interface between two indices
	Insert(i int, array Interface) Interface // Insert slice at position i
}

// A point basically returns coordinates
type Point interface {
	GetCoordinates() (float64, float64)
}

// Create an RBush index from an array of points
func New(points Interface) RBush {
	return RBush{}
}

type RBush struct {
	rootNode *Node
}

type Node struct {
	children           []*Node
	height int
	isLeaf             bool
	points             Interface
	parentNode         *Node
	bbox               BBox
}

func (r *RBush) Search(b BBox) []*Node {
	node := r.rootNode
	result := make([]*Node, 0)
	if !node.bbox.intersects(b) {
		return result
	}
	nodesToSearch := make([]*Node, 0, 1)
	nodesToSearch[1] = node
	for len(nodesToSearch) != 0 {
		// pop first item
		node, nodesToSearch = nodesToSearch[0], nodesToSearch[1:]
		for _, c := range(node.children) {
			if b.intersects(c.bbox) {
				if node.isLeaf {
					// child is basically a point
					result = append(result, c)
				} else if (b.contains(c.bbox)) {
					result = append(result, c.flattenDownwards())
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
	if !node.bbox.intersects(b) {
		return false
	}
	nodesToSearch := make([]*Node, 0, 10)
	nodesToSearch[0] = node
	for len(nodesToSearch) != 0 {
		// pop first item
		node, nodesToSearch = nodesToSearch[0], nodesToSearch[1:]
		for _, c := range(node.children) {
			// TODO leaf
			if c.bbox.intersects(b) {
				if (node.isLeaf || b.contains(c.bbox)) {
					return true
				}
				nodesToSearch = append(nodesToSearch, c)
			}
		}
	}
	return false
}

// Returns all end points inside node
func (n * Node) flattenDownwards () []*Node {
	var node *Node
	result := make([]*Node, 0, len(n.children))
	nodesToSearch := []*Node{n}
	for len(nodesToSearch) != 0 {
		node, nodesToSearch = nodesToSearch[0], nodesToSearch[1:]
		if node.isLeaf {
			result = append(result, node.children)
		} else {
			nodesToSearch = append(nodesToSearch, node.children)
		}
	}
	return result
}

func (r *RBush) Load(points Interface) {
	sort.Sort(pointSorter{i: points})
	r.LoadSortedArray(points)
}

func (r *RBush) LoadSortedArray(points Interface) {
	// TODO points.Len < MIN_ENTRIEs
	node := r.build(points)
	if r.rootNode == nil {
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
}

// points is assumed to be ordered
func (r *RBush) build(points Interface) Node {
	ch := make(chan *Node)
	readCh := make(chan *Node)
	exitCh := make(chan int, NUMBER_OF_SORTERS)
	for i := 0; i < NUMBER_OF_SORTERS; i++ {
		go func(ch, readCh chan *Node, exitCh chan int) {
			for true {
				select {
				case n := <-ch:
					N := n.points.Len() + 1 // why + 1 ?
					// target number of root entries to maximize storage utilization
					var M float64
					if N <= MAX_ENTRIES { // Leaf node
						n = r.createLeafNode(n.points)
						readCh <- n
						continue
					}
					// sort on x, then split in equal size buckets and sort in y
					// root node is assumed sorted so there is no need to sort
					// first node inserted
					if n.height == -1 {
						// This is the target height
						n.height = math.Ceil(math.Log(N) / math.Log(MAX_ENTRIES))
					} else {
						sortX := xSorter{n: n, start: 0, end: n.points.Len()}
						sort.Sort(sortX)
					}
					M = math.Ceil(float64(N) / float64(math.Pow(MAX_ENTRIES, n.height-1)))

					N2 := math.Ceil(N / M)
					N1 := N2 * math.Ceil(math.Sqrt(M))
					for i := 0; i <= n.points.Len(); i += N1 {
						right2 := math.Min(i+N-1, n.points.Len())
						sortY := ySorter{n: n, start: i, end: right2}
						sort.Sort(sortY)
						for j := i; j <= right2; j += N2 {
							right3 := math.Min(j+N2-1, right2)
							child := Node{
								points: n.points.Slice(j, right3),
								height: n.height - 1,
								parentNode: n,
							}
							n.children = append(n.children, &child)
							readCh <- &child
						}
					}
					// remove reference to interface, we only need it for points
					n.points = nil

				case <-exitCh:
					return
				}
			}

		}(ch, readCh, exitCh)
	}

	rootNode := Node{height: -1, points: points}
	ch <- &rootNode
	remainingNodes := 1

	for remainingNodes > 0 {
		select {
		case n := <-readCh:
			remainingNodes -= 1
			if n.isLeaf {
				continue // children of leaf nodes are just points so we should not try to create nodes out of there
			}
			for _, childNode := range n.children {
				remainingNodes += 1
				ch <- childNode
			}
		}
	}
	for i := 0; i < NUMBER_OF_SORTERS; i++ {
		exitCh <- 1
	}
	rootNode.computeBBoxDownwards()
	return rootNode
}

func (r *RBush) insertElement(p Interface) {
	x, y := p.GetCoordinatesAt(0)
	node := Node{
		points: p,
		bbox: BBox{
			MinX: x,
			MaxX: x,
			MinY: y,
			MaxY: y,
		},
	}
	// TODO make sure this actually works
	r.insertNode(node)
}

func (r *RBush) insertNode(n Node) {
	chosenNode := r.choseSubtree(n)
	// TODO probably do something in the case : n.isLeaf, chosenNode.isLeaf
	n.parentNode = &chosenNode
	chosenNode.children = append(chosenNode.children, &n)
	chosenNode.bbox = chosenNode.bbox.extend(n.bbox)

	// split on node overflow, propagate upwards
	for iterNode := chosenNode; iterNode != nil; iterNode = iterNode.parentNode {
		if len(iterNode.children) > MAX_ENTRIES {
			r.split(iterNode)
		} else {
			iterNode.bbox = iterNode.bbox.extend(n.bbox)
		}
	}

}

func (r *RBush) splitRoot(n Node) {
	newRoot := Node{
		height: r.rootNode.height + 1,
		children: []*Node{
			r.rootNode,
			&n,
		},
	}
	r.rootNode.parentNode = &newRoot
	n.parentNode = &newRoot
	r.rootNode = &newRoot
}

func (r *RBush) createLeafNode(p Interface) *Node {
	// Here we follow original rbush implementation.
	// TODO try to store elements children as points instead of nodes
	// It seems a bit inefficient to have one child for each point, but otherwise the complexity of the code blows up
	children := make([]*Node, p.Len())
	n := Node{
		children: children,
		isLeaf: true,
		height: 0,
	}

	for i:= 0; i < p.Len(); i++ {
		x, y := p.GetCoordinatesAt(i)
		children[i] = &Node{
			points: p.Slice(i, i+1),
			bbox: BBox{
				MinX: x,
				MaxX: x,
				MinY: y,
				MaxY: y,
			},
			parentNode: &n,
		}
	}
	return &n
}

// split node into two, update bboxes
func (r *RBush) split(n *Node) {
	n.chooseSplitAxis()
	i := n.chooseSplitIndex()
	newNode := Node{
		children: n.children[i: len(n.children) - 1],
		height: n.height,
		parentNode: n.parentNode,
		isLeaf: n.isLeaf,
	}
	n.children = n.children[0: i]
	for _, c := range(newNode.children) {
		c.parentNode = &newNode
	}
	n.bbox = n.partialBBox(0, len(n.children))
	newNode.bbox = newNode.partialBBox(0, len(newNode.children))
	// not root
	if n.parentNode != nil {
		n.parentNode.children = append(n.parentNode.children, newNode)
	} else {
		r.splitRoot(newNode)
	}

}

// sorts children by best axis for split
func (n *Node) chooseSplitAxis () {

}

// find best index to split
func (n *Node) chooseSplitIndex () int {

}

// find optimal node searching for the node that grows less in area.
func (r *RBush) choseSubtree(n Node) *Node {
	height := r.rootNode.height - n.height
	if height < 0 {
		// Most definitely an error in the implementation
		log.Fatal("We are inserting a big tree into a smaller tree.")
	}
	depth := 0
	chosenNode := r.rootNode
	for true {
		// We always insert small tree into big tree so it cannot happen that we insert a non point into a leaf
		if chosenNode.isLeaf || depth == height {
			break
		}
		minArea := math.MaxFloat64
		minEnlargement := math.MaxFloat64
		for _, child := range chosenNode.children {
			area := child.bbox.area()
			enlargement := n.bbox.enlargedArea(child.bbox) - area

			// find entry with minimum enlargment
			if enlargement < minEnlargement {
				minEnlargement = enlargement
				if area < minArea {
					minArea = area
				}
				chosenNode = child
			} else if enlargement == minEnlargement {
				if area < minArea {
					minArea = area
					chosenNode = child
				}
			}
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
			MinX: math.MaxFloat64,
			MaxX: -math.MaxFloat64,
			MinY: math.MaxFloat64,
			MaxY: -math.MaxFloat64,
		}
		// This bounded boxes are computed when creating the nodes, they only contain one point so there is no doubt
		for i:= 1; i < len(n.children); i ++ {
			bbox = bbox.extend(n.children[i].bbox)
		}
	} else {
		bbox = n.children[0].computeBBoxDownwards()
		for i:= 1; i < len(n.children); i ++ {
			bbox = bbox.extend(n.children[i].computeBBoxDownwards())
		}
	}
	n.bbox = bbox
	return bbox
}

// compute bbox of part of the childre
func (n *Node) partialBBox (start, end int) BBox {
	bbox := n.children[start].bbox
	for i := start + 1; i < end; i++ {
		bbox = bbox.extend(n.children[i].bbox)
	}
	return bbox
}

func (r RBush) Clear() {

}

func (r RBush) Remove() {

}

func (r RBush) ToBBox() {

}
