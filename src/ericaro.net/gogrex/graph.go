package gogrex

import (
	"fmt"
)

//Vertex represent any type that can act as a Vertex object
type Vertex interface{}

//Edge represent any type that can act as a transition between states
type Edge interface{}

type bounds struct {
	start, end Vertex
}

type DirectedSparseMultigraph struct {
	vertices map[Vertex]interface{}
	edges    map[Edge]bounds
}

func NewDirectedSparseMultigraph() *DirectedSparseMultigraph {
	return &DirectedSparseMultigraph{
		vertices: make(map[Vertex]interface{}),
		edges:    make(map[Edge]bounds),
	}
}

// remove the vertex state
func (g *DirectedSparseMultigraph) RemoveVertex(s Vertex) {
	for t, b := range g.edges {
		if b.end == s  || b.start == s {
			delete(g.edges, t)
		}
	}
	delete(g.vertices, s)
}
func (g *DirectedSparseMultigraph) AddVertex(s Vertex) {
	g.vertices[s] = nil
}

func (g *DirectedSparseMultigraph) RemoveEdge(t Edge) {
	delete(g.edges, t)
}

func (g *DirectedSparseMultigraph) AddEdge(t Edge, start, end Vertex) {
	g.edges[t] = bounds{start, end}
	g.AddVertex(start)
	g.AddVertex(end)
}

// 
func (g *DirectedSparseMultigraph) InEdges(s Vertex) (edges []Edge) {
	// a little bit brute force, but this is just the beginning
	for t, b := range g.edges {
		if b.end == s {
			edges = append(edges, t)
		}
	}
	return
}
func (g *DirectedSparseMultigraph) OutEdges(s Vertex) (edges []Edge) {
	// a little bit brute force, but this is just the beginning
	for t, b := range g.edges {
		if b.start == s {
			edges = append(edges, t)
		}
	}
	return
}

func (g *DirectedSparseMultigraph) Source(t Edge) Vertex {
	b := g.edges[t]
	return b.start
}
func (g *DirectedSparseMultigraph) Dest(t Edge) Vertex {
	b := g.edges[t]
	return b.end
}

func (g *DirectedSparseMultigraph) String(in Vertex, outs map[Vertex]interface{}) string {
	str := `digraph { size="6,4";rankdir=LR; ratio = fill; node [label="",shape=point,style=filled];
	`

	str += fmt.Sprintf(`%s [label="In",shape=box];
	`, in)
	for k := range outs {
		if k == in {
			str += fmt.Sprintf(`%s [label="IO %s",shape=box];
	`, k, k)
		} else {
			str += fmt.Sprintf(`%s [label="Out %s",shape=box];
	`, k, k)
		}
	}

	for t, b := range g.edges {
		str += fmt.Sprintf(`%s -> %s [label="%s"];
	`, b.start, b.end, t)
	}
	str += "}"
	return str
}
