package datalark_test

import (
	"fmt"
	"testing"

	"github.com/ipld/go-datalark/engine"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
)

func TestUsage(t *testing.T) {
	var n datamodel.Node
	var v datalarkengine.Value

	var dep *datalarkengine.Prototype
	var p ipld.NodePrototype

	dep = datalarkengine.NewPrototype("string", basicnode.Prototype.String)
	p = dep.NodePrototype()

	nb := p.NewBuilder()
	nb.AssignString("goodbye")
	n = nb.Build()

	v = datalarkengine.NewString("hello")

	fmt.Printf("%s\n", n)
	fmt.Printf("%v\n", n)
	fmt.Printf("%s\n", v)
	fmt.Printf("%v\n", v)
}
