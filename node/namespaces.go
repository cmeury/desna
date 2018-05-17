package node

import (
	"hash/fnv"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

type Namespaces struct {
	*simple.UndirectedGraph
	Sub []graph.Graph

}

func (g Namespaces) Structure() []graph.Graph {
	return g.Sub
}

//type namedGraph struct {
//	id int64
//	graph.Graph
//}
//
//func (g namedGraph) DOTID() string { return alpha[g.id : g.id+1] }

type Namespace struct {
	Name string
	graph.Graph
}

// ID returns the ID number of the node.
func (n Namespace) ID() int64 {
	new64a := fnv.New64a()
	new64a.Write([]byte(n.Name))
	return int64(new64a.Sum64())

}

func (g Namespace) Subgraph() graph.Graph {
	return Namespace(g)
}