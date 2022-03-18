package main

import (
	"errors"
	"fmt"
)

// -----------------------------------------------------------------------------

type Pair[U, V any] struct {
	First  U
	Second V
}

type Edge_[N comparable] struct {
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
}

func New_[N comparable, L any]() *Graph_[N, L] {
	graph := new(Graph_[N, L])
	graph.Nodes = map[N]bool{}
	graph.Edges = map[Edge_[N]]bool{}
	graph.Labels = map[Edge_[N]]L{}
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
	return fmt.Sprintf("nodes: %v\nedges: %v\nlabels: %v", graph.Nodes, graph.Edges, graph.Labels)
}

func (graph *Graph_[N, L]) Neighbors(node N) map[N]bool {
	neighbors := map[N]bool{}
	for edge, _ := range graph.Edges {
		if edge.Source == node {
			neighbors[edge.Target] = true
		}
	}
	return neighbors
}

func pop[K comparable](m map[K]bool) (K, error) {
	var k K
	if len(m) == 0 {
		return k, errors.New(fmt.Sprintf("empty map %v.\n", m))
	}
	for k = range m {
		break
	}
	delete(m, k)
	return k, nil
}

func (graph *Graph_[N, L]) PathTo(source, target N) []N {
	pathMap := map[N]([]N){source: []N{source}} // node -> path
	todo := map[N]bool{source: true}            // set of nodes
	done := map[N]bool{}                        // set of nodes
	for {
		node, err := pop(todo)
		if err != nil {
			break
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
	return nil
}

// -----------------------------------------------------------------------------

type Node = [2]int
type Edge = Edge_[Node]
type Label = float64
type Graph = Graph_[Node, Label]

var New func() *Graph = New_[Node, Label]

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
