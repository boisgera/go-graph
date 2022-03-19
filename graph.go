package main

import (
	"errors"
	"fmt"
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
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			nodes[Node{i, j}] = true
		}
	}
	todo := map[Node]bool{{0, 0}: true}
	done := map[Node]bool{}
LOOP:
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
					// TODO: clean-up
					continue LOOP
				}
			}
			// No neighbor was suitable here; need some clean-up and
			// continue the loop

		}
	}
}

// def dense_maze(width, height):
//     random.seed(0)
//     vertices = {(i, j) for i in range(width) for j in range(height)}
//     edges = set()
//     todo = {(0, 0)}  # visited but some neighbors not tested yet,
//     done = set()     # all neighbors have been tested.
//     while todo:
//         i, j = current = random.choice(list(todo))
//         neighbors = {(i + k, j + l) for k, l in [(1, 0), (0, 1), (-1, 0), (0, -1)]}
//         # neighbors in the maze and not explored yet
//         candidates = (neighbors & vertices) - done - todo
//         if candidates:
//             new = random.choice(list(candidates))
//             edges.add((current, new))
//             edges.add((new, current))  # both directions are allowed
//             todo.add(new)
//         if len(candidates) <= 1:
//             todo.remove(current)
//             done.add(current)
//     weights = {edge: 1.0 for edge in edges}
//     return vertices, edges, weights

func main() {
	graph := New()
	// fmt.Printf("%T\n", graph)
	// fmt.Printf("%#v %#v %#v\n", graph.Nodes, graph.Edges, graph.Labels)
	//fmt.Println(graph.Nodes == nil)
	graph.AddNode(Node{0, 0}, Node{1, 0}, Node{1, 1})
	graph.AddEdge(Edge{Node{0, 0}, Node{0, 1}}, Edge{Node{0, 1}, Node{1, 1}})
	fmt.Println(graph)
	path := graph.PathTo(Node{0, 0}, Node{1, 1})
	fmt.Println("path:", path)

}
