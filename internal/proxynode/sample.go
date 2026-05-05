package proxynode

import (
	"math/rand"
)

// SampleSubset shuffles a copy of nodes and returns up to k items (or all if fewer than k).
func SampleSubset(nodes []Node, k int) []Node {
	if k <= 0 || len(nodes) == 0 {
		return nil
	}
	copied := make([]Node, len(nodes))
	copy(copied, nodes)
	rand.Shuffle(len(copied), func(i, j int) {
		copied[i], copied[j] = copied[j], copied[i]
	})
	if len(copied) <= k {
		return copied
	}
	return copied[:k]
}
