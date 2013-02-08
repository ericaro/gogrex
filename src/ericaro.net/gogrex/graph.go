package gogrex

import (
	"fmt"
)

//Vertex represent any type that can act as a Vertex object
type Vertex interface{}

//Edge represent any type that can act as a transition between states
type Edge interface{
Name() string
}

//Bounds is a simple pair of start, and end Vertex
type Bounds struct {
	start, end Vertex
}

//Start return the Start of this Bound
func (b *Bounds) Start() Vertex{
	return b.start
}
//End return the End of this Bound
func (b *Bounds) End() Vertex{
	return b.end
}

//DirectedSparseMultigraph old a graph of (Vertex,Edge). Edges are directed, and every Vertex can have several inbounds, and outbounds
type DirectedSparseMultigraph struct {
	vertices map[Vertex]interface{}
	edges    map[Edge]Bounds
}

//NewDirectedSparseMultigraph creates a new empty graph
func NewDirectedSparseMultigraph() *DirectedSparseMultigraph {
	return &DirectedSparseMultigraph{
		vertices: make(map[Vertex]interface{}),
		edges:    make(map[Edge]Bounds),
	}
}

//RemoveVertex removes a vertex
func (g *DirectedSparseMultigraph) RemoveVertex(s Vertex) {
	for t, b := range g.edges {
		if b.end == s  || b.start == s {
			delete(g.edges, t)
		}
	}
	delete(g.vertices, s)
}
//AddVertex simply add an unconnected vertex to this graph
func (g *DirectedSparseMultigraph) AddVertex(s Vertex) {
	g.vertices[s] = nil
}

//RemoveEdge removes an edge from this graph. No Vertex is pruned
func (g *DirectedSparseMultigraph) RemoveEdge(t Edge) {
	delete(g.edges, t)
}

//AddEdge append a edge to the graph, and the start, and end vertex too.
func (g *DirectedSparseMultigraph) AddEdge(t Edge, start, end Vertex) {
	g.edges[t] = Bounds{start, end}
	g.AddVertex(start)
	g.AddVertex(end)
}

//InEdges computes the input Edges that reach 's' 
func (g *DirectedSparseMultigraph) InEdges(s Vertex) (edges []Edge) {
	// a little bit brute force, but this is just the beginning
	for t, b := range g.edges {
		if b.end == s {
			edges = append(edges, t)
		}
	}
	return
}
//OutEdges computes the output Edges that starts from 's'
func (g *DirectedSparseMultigraph) OutEdges(s Vertex) (edges []Edge) {
	// a little bit brute force, but this is just the beginning
	for t, b := range g.edges {
		if b.start == s {
			edges = append(edges, t)
		}
	}
	return
}

// note: there is a small discrepency between bounds (start,end) and Source/Dest naming

//Source returns the source of a given Edge
func (g *DirectedSparseMultigraph) Source(t Edge) Vertex {
	b := g.edges[t]
	return b.start
}


//Dest returns the Dest of a given Edge
func (g *DirectedSparseMultigraph) Dest(t Edge) Vertex {
	b := g.edges[t]
	return b.end
}
//String print this graph in dot format. 'in' usually the input vertex and outs, have different labels, and shape (box)
func (g *DirectedSparseMultigraph) String(in Vertex, outs map[Vertex]interface{}) string {
	str := `digraph { size="6,4";rankdir=LR; ratio = fill; node [label="",shape=point,style=filled];
	`

	str += fmt.Sprintf(`%s [label="In",shape=box];
	`, in)
	for k := range outs {
		if k == in {
			str += fmt.Sprintf(`%s [label="IO",shape=box];
	`, k)
		} else {
			str += fmt.Sprintf(`%s [label="Out",shape=box];
	`, k)
		}
	}

	for t, b := range g.edges {
		str += fmt.Sprintf(`%s -> %s [label="%s"];
	`, b.start, b.end, t.Name())
	}
	str += "}"
	return str
}
