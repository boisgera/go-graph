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

type Graph_[NodeType comparable, LabelType any] struct {
	Nodes  map[NodeType]bool
	Edges  map[Pair[NodeType, NodeType]]bool
	Labels map[Pair[NodeType, NodeType]]LabelType // would make more sense to have a single map
	// (less errors by construction, optimisation, etc.).
	// BUT there the conceptual design is clearer AND labels can be nil
	// (we don't "pay" for labels when there is not need).
}

func New_[NodeType comparable, LabelType any]() *Graph_[NodeType, LabelType] {
	graph := new(Graph_[NodeType, LabelType])
	graph.Nodes = map[NodeType]bool{}
	graph.Edges = map[Pair[NodeType, NodeType]]bool{}
	graph.Labels = map[Pair[NodeType, NodeType]]LabelType{}
	return graph
}

func (graph Graph_[N, L]) AddNode(nodes ...N) {
	for _, node := range nodes {
		graph.Nodes[node] = true
	}
}

func (graph Graph_[N, L]) AddEdge(edges ...Pair[N, N]) {
	for _, edge := range edges {
		graph.Edges[edge] = true
	}
}

func (graph Graph_[N, L]) String() string {
	return fmt.Sprintf("nodes: %v\nedges: %v\nlabels: %v", graph.Nodes, graph.Edges, graph.Labels)
}

func (graph *Graph_[N, L]) PathTo(source, destination N) ([]N, error) {
	var path []N
	message := fmt.Sprintf("no path from %v to %v", source, destination)
	err := errors.New(message)
	return path, err
}

// -----------------------------------------------------------------------------

type Node = int
type Edge = Pair[Node, Node]
type Label = string
type Graph = Graph_[Node, Label]

func New() *Graph { // no way to do this without an extra function call?
	return New_[Node, Label]()
}

func main() {
	graph := New()
	// fmt.Printf("%T\n", graph)
	// fmt.Printf("%#v %#v %#v\n", graph.Nodes, graph.Edges, graph.Labels)
	//fmt.Println(graph.Nodes == nil)
	graph.AddNode(1, 2, 3)
	graph.AddEdge(Edge{1, 2}, Edge{2, 3})
	fmt.Println(graph)
	_, err := graph.PathTo(1, 2)
	if err != nil {
		panic(err)
	}

}
