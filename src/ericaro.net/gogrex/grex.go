package gogrex

import (
	"errors"
	"fmt"
	"strings"
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

// a grex is not a regular graph, is a graph build by a regular expression

// manage state and transitions (creation and duplication)
// everyone interested in build a grex from a regular expression shall implement a Manager (sorry for the name)
type Manager interface {
	//Called to build a new vertex for the graph
	NewVertex() Vertex
	//build a new edge. The name is the symbol used in the regular expression
	NewEdge(name string) Edge
	//clone the edge, the same "edge" will be reused between other edges
	CloneEdge(t Edge) Edge
}

//String manager is a basic implementation, string oriented of a Manager
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


//Grex is a graph of a regular expression. 
type Grex struct {
	manager Manager // kind of metaclass, to build edges and vertice
	in      Vertex // every grex has a single input vertex
	outs    map[Vertex]interface{} //a grex has several outputs vertice. We use a map for that purpose

	graph *DirectedSparseMultigraph // the graph for real.
}

func (g *Grex) String() string {
	return g.graph.String(g.in, g.outs) // tmp
}

//NewGrex creates an empty grex based on a manager.
func NewGrex(m Manager) *Grex {
	return &Grex{
		outs:    make(map[Vertex]interface{}),
		graph:   NewDirectedSparseMultigraph(),
		manager: m,
	}
}

//terminal is called when parsing a indentifier, in charge to build a basic grex made of a single edge
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
//        |       | outs       	      |       | outs                    |       | outs       	      |       | outs                   	
//      in|       |-----       	    in|       |-----           	      in|       |---------------------|       |-----        
//      --|-this  |         ,       --|-that  |          =>           --|-this  |         ,           |-that  |           
//        |       |-----       	      |       |-----           	        |       |---------------------|       |-----        
func seq(this, that *Grex) *Grex {

	n := NewGrex(this.manager) // new empty  Grex

	// copyGraphInto clones this or that into n, an new grex. Edges, and vertices are cloned so they are not connected.
	// the map returned, maps from a source vertex to its target. 
	mapThis := this.copyGraphInto(n) 
	mapThat := that.copyGraphInto(n)
	
	_, io := that.outs[that.in] // io is true if the input vertex is also an output one ( its possible), think a*
	
	in := mapThat[that.in] // in contains the vertex that was that input. The idea is to connect all outputs of this to it
	 
	for out := range this.outs { // for every outs of this (the o  nexus )
		n.mergeOutbounds(in, mapThis[out])
		if io { //  "that"'s input is also an output  , therefore every output of "this" (out) should be appended to the outputs of the new grex  
			n.outs[mapThis[out]] = nil
		}
	}
	
	// new  "in" is this.in, and new  outs are that.outs
	n.in = mapThis[this.in]
	for out := range that.outs {
		n.outs[mapThat[out]] = nil
	}
	
	// "this".outs were not connected to "that".in. Instead, all the outbounds of "that".in (i.e "that".in.outs ) are copied to "this".outs
	//therefore the vertex in is no longer needed.
	//Pruning it 
	n.graph.RemoveVertex(in)
	delete(n.outs, mapThat[that.in])
	return n
}

//returns a new Grex result of  ( this |  that )
//        |       | outs       	      |       | outs                    	
//      in|       |-----       	    in|       |-----         
//      --|-this  |         or      --|-that  |            
//        |       |-----       	      |       |-----     
//
//  =>
//
//                              |       |  outs 
//                              |       |------------------- 
//                            --|-this  |      
//                in         |  |       |-------------------              
//                -----------                                                 
//                           |  |       |  
//                           |  |       |------------------- 
//                            --|-that  |      
//                              |       |------------------- 
//
//
//
//
//        
func sel(this, that *Grex) *Grex {

	n := NewGrex(this.manager)

	// clone "this", and "that", and keep a map to know who's who
	mapThis := this.copyGraphInto(n) 
	mapThat := that.copyGraphInto(n)

	// "this".in "that".in must be merged
	ms1 := mapThis[this.in]
	ms2 := mapThat[that.in]

	// merge them together into a new one
	n.in = n.mergeRaw(ms1, ms2)

	// replace the original items in the map, to be able to reference it again
	mapThis[this.in] = n.in
	mapThat[that.in] = n.in

	// now simply append all "that".outs to "new".outs
	for out := range that.outs {
		n.outs[mapThat[out]] = nil
	}
	// now simply append all "this".outs to "new".outs
	for out := range this.outs {
		n.outs[mapThis[out]] = nil
	}
	return n
}

//plus returns a new Grex result of  ( this )+
// this is simply adding a edge between all outputs of "this" back to "this".in
func plus(this *Grex) *Grex {
	// graph
	n := this.dup() // duplicate the graph
	//every outputs of n (n.outs) can reach exactly the same edges that n.in can.
	for out := range n.outs {
		n.mergeOutbounds(n.in, out)
	}
	return n
}

//opt returns a new Grex result of  ( this )?
// this is simply adding in as an output.
func opt(this *Grex) *Grex {
	n := this.dup()
	n.outs[n.in] = nil
	return n
}

//star returns a new Grex result of  ( this )*
// here we cheated, ()* is implemented as ()+?
func star(this *Grex) *Grex {
	return opt(plus(this))
}

// #################################################################################################
// Graph Operators
// #################################################################################################

//InputVertex returns the input vertex of this grex.
func (g *Grex) InputVertex() Vertex {
	return g.in
}
//OutputVertice return a copied slice of the outputs
func (g *Grex) OutputVertice() (outputs []Vertex) {
	for v := range g.outs {
		outputs = append(outputs, v)
	}
	return
}

//OutputEdges return all the outputs Edges of a given vertex
func (g *Grex) OutputEdges(vertex Vertex) (outputs []Edge) {
	return g.graph.OutEdges(vertex)
}

//VertexByPath returns the vertex denoted by a vertex path, "." separated, of every vertex name
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

//Vertices return a copied slice of all vertices
func (g *Grex) Vertices() (vertices []Vertex) {
	for v := range g.graph.vertices {
		vertices = append(vertices, v)
	}
	return
}

//Edges returns a map of all Edges, and their Bounds ( a simple pair (Source, Dest)
func (g *Grex) Edges() map[Edge]Bounds {
	return g.graph.edges
}

//mergeRaw creates a new vertex that has all the outbounds of both s1, and s2, and all their inbounds too.
func (g *Grex) mergeRaw(s1, s2 Vertex) Vertex {
	s := g.manager.NewVertex()

	// computes the global inbounds
	incomings := make(map[Edge]interface{})
	for _, t := range g.graph.InEdges(s1) {
		incomings[t] = nil
	}
	for _, t := range g.graph.InEdges(s2) {
		incomings[t] = nil
	}
	
	// each incoming edge is moved from source to s1|s2 to the new one.
	for t := range incomings {
		src := g.graph.Source(t) // the source of the incoming
		g.graph.RemoveEdge(t) // the edge is removed first, 
		g.graph.AddEdge(t, src, s) // then reappended with the new bounds
	}
	// proceed the same for outbounds
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
		g.graph.AddEdge(t, s, dst)
	}
	// now prune the removed s1, and s2
	g.graph.RemoveVertex(s1)
	g.graph.RemoveVertex(s2)
	return s
}

//copyGraphInto simply clone this all transition and vertices in the target graph
//and returns a map that link vertices from the source to vertices to the copy
func (g *Grex) copyGraphInto(target *Grex) map[Vertex]Vertex {
	m := make(map[Vertex]Vertex)

	// clone all vertices, and store in the map
	for s := range g.graph.vertices {
		c := target.manager.NewVertex() // clone
		m[s] = c                        // kept for transition clone
	}
	// clone all edges, and append to the graph
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

//dup simply duplicates a grex
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

//ParseGrex parses the regexp, and build a new Grex, using the Manager
func ParseGrex(m Manager, regexp string) (grex *Grex, err error) {
	tokens := lex(regexp) // build a lexer
	grammar, errchan := shunting(tokens) // start the shuntingYard
	
	// now parses the expression in a RPN notation
	var stack grexStack  // as any RPN interpreter I need a stack
	var t Token
	for {
		t, err = nil, nil
		select { // that's a bit ugly to my taste
		case t = <-grammar:
		case err = <-errchan:
		} //get a correct token, or an error
		if t == nil && err == nil { // end detected
			this, err := stack.Pop() // return the last item in the stack
			if err != nil {
				return nil, err
			}
			return this, nil
		}
		if err != nil { // not the end, but an error though
			return
		}
		i := t.(item) // now I've got an item
		switch i.typ { // operates,
		case itemStar: // mono operand: pop, star
			this, err := stack.Pop()
			if err != nil {
				return nil, err
			}
			this = star(this)
			stack.Push(this)
		case itemPlus: // mono operand : pop, plus
			this, err := stack.Pop()
			if err != nil {
				return nil, err
			}
			this = plus(this)
			stack.Push(this)
		case itemOpt: //mono operand : pop, star
			this, err := stack.Pop()
			if err != nil {
				return nil, err
			}
			this = opt(this)
			stack.Push(this)
		case itemSel: // binary operand: pop, pop, sel
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
		case itemSeq:// binary operand: pop, pop, sel
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
		case itemIdentifier:// leaf element, creates a new single edge graph
			this := terminal(m, i.val)
			stack.Push(this)
		case itemError: // a lex error has occured
			err = errors.New(i.val)
			return
		default: // unexpected token
			err = errors.New(fmt.Sprintf("Invalid Token %s.", i.val))
			return
		}
	}
	return
}
