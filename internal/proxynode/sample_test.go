package proxynode

import (
	"fmt"
	"testing"
)

func TestSampleSubset_TakesAtMostK(t *testing.T) {
	nodes := make([]Node, 15)
	for i := range nodes {
		nodes[i] = Node{Protocol: "ss", Host: "h", Port: i, RawURI: fmt.Sprintf("ss://x-%d", i)}
	}
	out := SampleSubset(nodes, 10)
	if len(out) != 10 {
		t.Fatalf("len %d", len(out))
	}
	seen := map[string]bool{}
	for _, n := range out {
		if seen[n.RawURI] {
			t.Fatal("duplicate")
		}
		seen[n.RawURI] = true
	}
}

func TestSampleSubset_AllWhenFewer(t *testing.T) {
	nodes := []Node{{RawURI: "a"}, {RawURI: "b"}}
	out := SampleSubset(nodes, 10)
	if len(out) != 2 {
		t.Fatalf("len %d", len(out))
	}
}
