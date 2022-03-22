package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"image/color"
	"strings"

	"github.com/llgcode/draw2d/draw2dsvg"
)

// -----------------------------------------------------------------------------
type Edge_[N comparable] struct { // bit overkill, use [2]Node instead?
	Source N
	Target N
}

type Graph_[N comparable, L any] struct {
	Nodes  map[N]bool
	Edges  map[Edge_[N]]bool
	Labels map[Edge_[N]]L // would make more sense to have a single map
	// (less errors by construction, optimisation, etc.).
	// BUT there the conceptual design is clearer AND labels can be nil
	// (we don't "pay" for labels when there is not need).
	// Well ok but then Labels should be external and maybe we could compound
	// both stuff into a new struct ("inherit")
}

func New_[N comparable, L any]() *Graph_[N, L] {
	graph := new(Graph_[N, L])
	graph.Nodes = map[N]bool{}
	graph.Edges = map[Edge_[N]]bool{}
	graph.Labels = map[Edge_[N]]L{} // or nil?
	return graph
}

func (graph Graph_[N, L]) AddNode(nodes ...N) {
	for _, node := range nodes {
		graph.Nodes[node] = true
	}
}

func (graph Graph_[N, L]) AddEdge(edges ...Edge_[N]) {
	for _, edge := range edges {
		graph.Edges[edge] = true
	}
}

func (graph Graph_[N, L]) String() string {
	return fmt.Sprintf(
		"nodes: %v\nedges: %v\nlabels: %v",
		graph.Nodes,
		graph.Edges,
		graph.Labels,
	)
}

func (graph *Graph_[N, L]) Neighbors(node N) map[N]bool {
	neighbors := map[N]bool{}
	for edge := range graph.Edges {
		if edge.Source == node {
			neighbors[edge.Target] = true
		}
	}
	return neighbors
}

// Pop an elt from a set.
func pop[E comparable](set map[E]bool) (E, error) {
	var elt E
	var err error

	if len(set) == 0 {
		message := fmt.Sprintf("Can't pop element from empty set %v.\n", set)
		err = errors.New(message)
		return elt, err
	}

	// Start a set iteration, but stop at the first element.
	for elt = range set {
		break
	}
	delete(set, elt)
	return elt, nil
}

func (graph *Graph_[N, L]) PathTo(source, target N) []N {
	pathMap := map[N]([]N){source: []N{source}} // node -> path
	todo := map[N]bool{source: true}            // set of nodes
	done := map[N]bool{}                        // set of nodes
	for {
		node, err := pop(todo)
		if err != nil { // todo is empty, no path found.
			return nil
		}

		path := pathMap[node]
		for neighbor := range graph.Neighbors(node) {
			if done[neighbor] {
				continue
			}
			newPath := make([]N, len(path), len(path)+1)
			copy(newPath, path)
			newPath = append(newPath, neighbor)
			if neighbor == target {
				return newPath
			}
			pathMap[neighbor] = newPath
			todo[neighbor] = true
		}
		done[node] = true
	}
}

// -----------------------------------------------------------------------------

type Node = [2]int
type Edge = Edge_[Node]
type Label = float64
type Graph = Graph_[Node, Label]

var New func() *Graph = New_[Node, Label]

// We don't do random in the go version, but some pseudo-randomness comes
// from the use of maps to contain nodes.
func NewDenseMaze(width, height int) *Graph {
	maze := New()
	nodes := map[Node]bool{}
	maze.Nodes = nodes
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			nodes[Node{i, j}] = true
		}
	}
	todo := map[Node]bool{{0, 0}: true}
	done := map[Node]bool{}

	for {
		node, err := pop(todo)
		if err != nil { // todo is empty, job done! 🥳
			return maze
		}

		i, j := node[0], node[1]
		deltas := [4]Node{{-1, 0}, {0, -1}, {1, 0}, {0, 1}}
		neighbors := map[Node]bool{}
		for _, delta := range deltas {
			n := Node{i + delta[0], j + delta[1]}
			if nodes[n] {
				neighbors[n] = true
			}

			for n := range neighbors {
				if nodes[n] && !done[n] && !todo[n] {
					maze.AddEdge(Edge{node, n}, Edge{n, node})
					todo[n] = true
				}
			}
			done[node] = true
			delete(todo, node)
		}
	}
}

type Points map[[2]int]bool // well, we would need to use that in our structure ...
// Just to please JSON. I'd rather make my own json serializer ATM.

func (points Points) MarshalJSON() ([]byte, error) {
	var bytes []byte
	bytes = append(bytes, '[')
	for point := range points {
		str := fmt.Sprintf("[%d, %d]", point[0], point[1])
		bytes = append(bytes, []byte(str)...)
	}
	bytes = append(bytes, ']')
	return bytes, nil
}

func toJSON(graph *Graph) string {
	var buffer []string

	buffer = append(buffer, `{"nodes": `)
	buffer = append(buffer, "[")
	for node := range graph.Nodes {
		s := fmt.Sprintf("[%d, %d],", node[0], node[1])
		buffer = append(buffer, s)
	}
	// Remove the trailing comma.
	last := buffer[len(buffer)-1]
	buffer[len(buffer)-1] = last[:len(last)-1]
	buffer = append(buffer, "],")

	buffer = append(buffer, ` "edges": [`)
	for edge := range graph.Edges {
		source := edge.Source
		target := edge.Target
		s := fmt.Sprintf(
			"[[%d, %d], [%d, %d]],",
			source[0],
			source[1],
			target[0],
			target[1],
		)
		buffer = append(buffer, s)
	}
	// Remove the trailing comma.
	last = buffer[len(buffer)-1]
	buffer[len(buffer)-1] = last[:len(last)-1]

	buffer = append(buffer, "]")
	buffer = append(buffer, "}")

	return strings.Join(buffer, "")

}

func min(ints ...int) int {
	m := ints[0]
	for _, i := range ints {
		if i < m {
			m = i
		}
	}
	return m
}

func max(ints ...int) int {
	m := ints[0]
	for _, i := range ints {
		if i > m {
			m = i
		}
	}
	return m
}

func SvgToBytes(svg *draw2dsvg.Svg) []byte {
	// f, err := os.Create(filePath)
	// if err != nil {
	// 	return err
	// }
	// defer f.Close()

	var buffer bytes.Buffer

	// create a file-like stuff instead that we can get a string from.

	buffer.Write([]byte(xml.Header))
	encoder := xml.NewEncoder(&buffer)
	encoder.Indent("", "\t")
	_ = encoder.Encode(svg)

	return buffer.Bytes()
}

func drawMaze(maze *Graph, width, height, widthMargin, heightMargin int, filename string) []byte {
	// Initialize the graphic context on an RGBA image

	scale := 10

	svg := draw2dsvg.NewSvg()
	svg.Width = fmt.Sprintf("%d", width*scale)
	svg.Height = fmt.Sprintf("%d", height*scale)
	svg.ViewBox = fmt.Sprintf("0 0 %d %d", width*scale, height*scale)

	nodes := maze.Nodes
	is := make([]int, len(nodes))
	js := make([]int, len(nodes))
	k := 0
	for node := range nodes {
		is[k] = node[0]
		js[k] = node[1]
		k += 1
	}
	iMin := min(is...)
	iMax := max(is...) + 1
	jMin := min(js...)
	jMax := max(js...) + 1
	println(iMin, iMax, jMin, jMax)

	dx := float64(width-2*widthMargin) / float64(iMax-iMin) * float64(scale)
	dy := float64(height-2*heightMargin) / float64(jMax-jMin) * float64(scale)

	xy := func(i, j int) (float64, float64) { // return the lower cell corner
		return float64(i)*dx + float64(widthMargin*scale), float64(j)*dy + float64(heightMargin*scale)
	}

	c := draw2dsvg.NewGraphicContext(svg)

	// Set some properties
	//c.SetFillColor(color.RGBA{0x44, 0xff, 0x44, 0xff})
	c.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
	c.SetLineWidth(1)

	edges := maze.Edges
	for node := range nodes {
		i := node[0]
		j := node[1]
		x, y := xy(i, j)
		if !edges[Edge{[2]int{i, j}, [2]int{i, j - 1}}] {
			c.BeginPath()  // Initialize a new path
			c.MoveTo(x, y) // Move to a position to start the new path
			c.LineTo(x+dx, y)
			c.Close()
			c.FillStroke()
		}
		if !edges[Edge{[2]int{i, j}, [2]int{i + 1, j}}] {
			c.BeginPath()     // Initialize a new path
			c.MoveTo(x+dx, y) // Move to a position to start the new path
			c.LineTo(x+dx, y+dy)
			c.Close()
			c.FillStroke()
		}
		if !edges[Edge{[2]int{i, j}, [2]int{i, j + 1}}] {
			c.BeginPath()        // Initialize a new path
			c.MoveTo(x+dx, y+dy) // Move to a position to start the new path
			c.LineTo(x, y+dy)
			c.Close()
			c.FillStroke()
		}
		if !edges[Edge{[2]int{i, j}, [2]int{i - 1, j}}] {
			c.BeginPath()     // Initialize a new path
			c.MoveTo(x, y+dy) // Move to a position to start the new path
			c.LineTo(x, y)
			c.Close()
			c.FillStroke()
		}
		// Draw a closed shape
		// c.BeginPath()    // Initialize a new path
		// c.MoveTo(10, 10) // Move to a position to start the new path
		// c.LineTo(100, 50)
		// c.QuadCurveTo(100, 10, 10, 10)
		// c.Close()
		// c.FillStroke()
	}

	// Save to file
	//draw2dsvg.SaveToSvgFile(filename, svg)
	return SvgToBytes(svg)
}

func main() {
	// graph := New()
	// graph.AddNode(Node{0, 0}, Node{1, 0}, Node{1, 1})
	// graph.AddEdge(Edge{Node{0, 0}, Node{0, 1}}, Edge{Node{0, 1}, Node{1, 1}})
	// fmt.Println(graph)
	// path := graph.PathTo(Node{0, 0}, Node{1, 1})
	// fmt.Println("path:", path)

	// fmt.Println("--------------------------------------------------")
	width, height := 40, 30
	maze := NewDenseMaze(width, height)
	fmt.Println(maze)

	// -----------------------------------------------------------------
	// JSONString := toJSON(maze)
	// file, err := os.Create(fmt.Sprintf("maze-%dx%d.json", width, height))
	// if err != nil {
	// 	panic(err)
	// }
	// defer file.Close()
	// file.Write([]byte(JSONString))

	// -----------------------------------------------------------------------
	fmt.Println("--------------------------------------------------------------")
	bytes := drawMaze(maze, width, height, 1, 1, fmt.Sprintf("maze-%dx%d.svg", width, height))
	fmt.Println(string(bytes))

	<-make(chan int)
}
