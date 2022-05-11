package datalarkengine

import (
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("%v != %v", a, b)
	}
}

func TestBasicTypes(t *testing.T) {
	var val Value

	val = NewNull()
	assertEqual(t, val.String(), "null")
	assertEqual(t, val.Type(), "datalark.null")

	val = NewBool(true)
	assertEqual(t, val.String(), "bool{true}")
	assertEqual(t, val.Type(), "datalark.bool")

	val = NewInt(34)
	assertEqual(t, val.String(), "int{34}")
	assertEqual(t, val.Type(), "datalark.int")

	val = NewFloat(7.2)
	// TODO(dustmop): Why doesn't this render the float value?
	assertEqual(t, val.String(), "float{}")
	assertEqual(t, val.Type(), "datalark.float")

	val = NewString("hi")
	assertEqual(t, val.String(), "string{\"hi\"}")
	assertEqual(t, val.Type(), "datalark.string")

	val = NewBytes([]byte{0x12, 0x56, 0x90})
	assertEqual(t, val.String(), "bytes{125690}")
	assertEqual(t, val.Type(), "datalark.bytes")

	val = NewLink(newTestLink())
	assertEqual(t, val.String(), "link{bafkqabiaaebagba}")
	assertEqual(t, val.Type(), "datalark.link")
}

func newTestLink() datamodel.Link {
	// Example link from:
	// https://github.com/ipld/go-ipld-prime/blob/master/datamodel/equal_test.go
	someCid, _ := cid.Cast([]byte{1, 85, 0, 5, 0, 1, 2, 3, 4})
	return cidlink.Link{Cid: someCid}
}
