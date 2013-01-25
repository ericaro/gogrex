package gogrex

import (
	"errors"
	"fmt"
	"strings"
)

var (
	_ = fmt.Print
)

type grexStack []*Grex

func (stack *grexStack) Pop() (i *Grex, err error) {
	i, err = stack.Peek()
	if err == nil {
		*stack = (*stack)[:len(*stack)-1]
	}
	return
}

func (stack *grexStack) Push(x *Grex) {
	*stack = append(*stack, x)
}

func (stack *grexStack) Peek() (i *Grex, err error) {
	if len(*stack) == 0 {
		return i, errors.New("Empty Stack")
	}
	return (*stack)[len(*stack)-1], nil
}

// a grex is a regular graph. What is remarquable is that it can be defined by a regular expression (hence their common name)

// manage state and transitions (creation and duplication)
type Manager interface {
	NewVertex() Vertex
	NewEdge(name string) Edge
	CloneEdge(t Edge) Edge
}

// manage state and transitions (creation and duplication)
type StringManager int
type trans struct {
	id  int
	str string
}

func (t trans) String() string {
	return fmt.Sprintf("%s(%d)", t.str, t.id)
}
func (t trans) Name() string {
	return t.str
}

func (m *StringManager) NewVertex() Vertex {
	*m = *m + 1
	return fmt.Sprintf("%d", *m)
}
func (m *StringManager) NewEdge(name string) Edge {
	*m = *m + 1
	return &trans{int(*m), name}
}
func (m *StringManager) CloneEdge(e Edge) Edge {
	t := e.(*trans)
	*m = *m + 1
	return &trans{int(*m), t.str}
}

type Grex struct {
	manager Manager
	in      Vertex
	outs    map[Vertex]interface{}

	graph *DirectedSparseMultigraph
	// plus what really make it a graph

}

func (g *Grex) String() string {
	return g.graph.String(g.in, g.outs) // tmp
}

func NewGrex(m Manager) *Grex {
	return &Grex{
		outs:    make(map[Vertex]interface{}),
		graph:   NewDirectedSparseMultigraph(),
		manager: m,
	}
}

func terminal(m Manager, name string) *Grex {
	g := NewGrex(m)
	t := m.NewEdge(name)
	s1 := m.NewVertex()
	s2 := m.NewVertex()
	g.graph.AddEdge(t, s1, s2)
	g.in = s1
	g.outs[s2] = nil
	return g
}

// return a new grex result of "this,that"  
func seq(this, that *Grex) *Grex {

	n := NewGrex(this.manager) // new empty  Grex

	// graph operations  // I still don't know what a graph looks like in go
	mapThis := this.copyGraphInto(n) // return an association map
	mapThat := that.copyGraphInto(n)

	//fmt.Printf("seq raw:  %s\n", n.String())
	// if 
	_, io := that.outs[that.in]

	in := mapThat[that.in]
	// copy 
	for out := range this.outs {
		n.mergeOutbounds(in, mapThis[out])
		if io {
			n.outs[mapThis[out]] = nil
		}
	}
	//fmt.Printf("seq merged in:  %s\n", n.String())

	// result in is this.in, and resutl outs are that outs
	n.in = mapThis[this.in]
	for out := range that.outs {
		n.outs[mapThat[out]] = nil
	}

	n.graph.RemoveVertex(in)
	delete(n.outs, mapThat[that.in])
	return n
}

//returns a new Grex result of  ( this |  that )
func sel(this, that *Grex) *Grex {

	n := NewGrex(this.manager)

	// graph operations
	mapThis := this.copyGraphInto(n) // return an association map
	mapThat := that.copyGraphInto(n)

	//bounds
	//n.in = n.merge(mapThis, this.in, mapThat, that.in); // inlined

	// get the both versions of the same node
	ms1 := mapThis[this.in]
	ms2 := mapThat[that.in]

	// merge them together into a new one
	n.in = n.mergeRaw(ms1, ms2)

	// replace the original items in the map
	mapThis[this.in] = n.in
	mapThat[that.in] = n.in

	for out := range that.outs {
		n.outs[mapThat[out]] = nil
	}
	for out := range this.outs {
		n.outs[mapThis[out]] = nil
	}
	return n
}

//plus returns a new Grex result of  ( this )+
func plus(this *Grex) *Grex {
	// graph
	n := this.dup()
	//bounds
	for out := range n.outs {
		n.mergeOutbounds(n.in, out)
	}
	return n
}

//opt returns a new Grex result of  ( this )?
func opt(this *Grex) *Grex {
	n := this.dup()
	n.outs[n.in] = nil
	return n
}

//star returns a new Grex result of  ( this )*
func star(this *Grex) *Grex {
	return opt(plus(this))
}

// #################################################################################################
// Graph Operators
// #################################################################################################

func (g *Grex) InputVertex() Vertex {
	return g.in
}

func (g *Grex) OutputVertice() (outputs []Vertex) {
	for v := range g.outs {
		outputs = append(outputs, v)
	}
	return
}

func (g *Grex) OutputEdges(vertex Vertex) (outputs []Edge) {
	return g.graph.OutEdges(vertex)
}

//VertexByPath returns the state by a vertex path, "." separated
func (g *Grex) VertexByPath(path string) Vertex {

	elements := strings.Split(path, ".")
	current := g.in

	for _, next := range elements {
		for _, t := range g.graph.OutEdges(current) {
			if next == t.Name() {
				current = g.graph.Dest(t)
				break // get out of the transition loop
			}
		}
	}
	// now the whole path has been parsed, and current is our goal
	return current
}

//Vertices return a slice of actual vertices
func (g *Grex) Vertices() (vertices []Vertex) {
	for v := range g.graph.vertices {
		vertices = append(vertices, v)
	}
	return
}

//Edges returns the state by a vertex path, "." separated
func (g *Grex) Edges() map[Edge]Bounds {
	return g.graph.edges
}

func (g *Grex) mergeRaw(s1, s2 Vertex) Vertex {
	s := g.manager.NewVertex()

	incomings := make(map[Edge]interface{})
	for _, t := range g.graph.InEdges(s1) {
		incomings[t] = nil
	}
	for _, t := range g.graph.InEdges(s2) {
		incomings[t] = nil
	}

	for t := range incomings {
		src := g.graph.Source(t)
		g.graph.RemoveEdge(t)
		g.graph.AddEdge(t, src, s)
	}
	outcomings := make(map[Edge]interface{})
	for _, t := range g.graph.OutEdges(s1) {
		outcomings[t] = nil
	}
	for _, t := range g.graph.OutEdges(s2) {
		outcomings[t] = nil
	}
	for t := range outcomings {
		dst := g.graph.Dest(t)
		g.graph.RemoveEdge(t)
		fmt.Printf("adding edge %v: %v->%v\n", t, s, dst)
		g.graph.AddEdge(t, s, dst)

	}

	g.graph.RemoveVertex(s1)
	g.graph.RemoveVertex(s2)
	return s
}

//copyGraphInto simply clone this all transition and vertices in the target graph
func (g *Grex) copyGraphInto(target *Grex) map[Vertex]Vertex {
	m := make(map[Vertex]Vertex)

	for s := range g.graph.vertices {
		c := target.manager.NewVertex() // clone
		m[s] = c                        // kept for transition clone
		//target.graph.AddVertex(c)       // append to the new graph
	}
	for t, b := range g.graph.edges {
		tclone := target.manager.CloneEdge(t)
		target.graph.AddEdge(tclone, m[b.start], m[b.end])
	}
	return m
}

//mergeOutbounds copies src outbounds into dest ones.
func (g *Grex) mergeOutbounds(src, dest Vertex) {
	// copy oldout  outbonds into source
	for t, bounds := range g.graph.edges {
		if bounds.start == src {
			g.graph.AddEdge(g.manager.CloneEdge(t), dest, bounds.end)
		}
	}
}

func (g *Grex) dup() *Grex {
	that := NewGrex(g.manager)
	// graph
	m := g.copyGraphInto(that)
	//bounds
	that.in = m[g.in]
	for out := range g.outs {
		that.outs[m[out]] = nil
	}
	return that
}

func (g *Grex) cloneEdge(t Edge) Edge {
	return g.manager.CloneEdge(t)
}

// parse is a temporary method to display
func ParseGrex(m Manager, str string) (grex *Grex, err error) {
	tokens := lex(str)
	grammar, errchan := shunting(tokens)
	var stack grexStack
	var t Token
	for {
		t, err = nil, nil
		select {
		case t = <-grammar:
		case err = <-errchan:
		}
		if t == nil && err == nil { // end detected
			//fmt.Printf("this is the end\n")
			this, err := stack.Pop()
			if err != nil {
				return nil, err
			}
			return this, nil
		}
		if err != nil {
			return
		}
		i := t.(item)
		//fmt.Printf("Processing %s\n", i.val)
		switch i.typ {
		case itemStar:
			this, err := stack.Pop()
			if err != nil {
				return nil, err
			}
			this = star(this)
			stack.Push(this)
		case itemPlus:
			this, err := stack.Pop()
			if err != nil {
				return nil, err
			}
			this = plus(this)
			stack.Push(this)
		case itemOpt:
			this, err := stack.Pop()
			if err != nil {
				return nil, err
			}
			this = opt(this)
			stack.Push(this)
		case itemSel:
			b, err := stack.Pop()
			if err != nil {
				return nil, err
			}
			a, err := stack.Pop()
			if err != nil {
				return nil, err
			}
			this := sel(a, b)
			stack.Push(this)
		case itemSeq:
			b, err := stack.Pop()
			if err != nil {
				return nil, err
			}
			a, err := stack.Pop()
			if err != nil {
				return nil, err
			}
			this := seq(a, b)
			stack.Push(this)
		case itemIdentifier:
			this := terminal(m, i.val)
			stack.Push(this)
		case itemError:
			err = errors.New(i.val)
			return
		default:
			err = errors.New(fmt.Sprintf("Invalid Token %s.", i.val))
			return
		}
	}
	return
}
