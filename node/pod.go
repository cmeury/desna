package node

import (
	"hash/fnv"
)

type Pod struct {
	Name      string
	Namespace string
}

// ID returns the ID number of the node.
func (n Pod) ID() int64 {
	new64a := fnv.New64a()
	new64a.Write([]byte(n.Namespace + n.Name))
	return int64(new64a.Sum64())

}

func (n Pod) DOTID() string {
	return "\"" + n.Name + "\""
}