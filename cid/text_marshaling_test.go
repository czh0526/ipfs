package cid

import (
	"fmt"
	"testing"

	gcid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

func TestTextMarshaling(t *testing.T) {
	data := []byte("this is some test content")
	hash, _ := mh.Sum(data, mh.SHA2_256, -1)
	c := gcid.NewCidV1(gcid.DagCBOR, hash)
	var c2 gcid.Cid

	data, err := c.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("marshal text ==> %s \n", data)

	err = c2.UnmarshalText(data)
	if err != nil {
		t.Fatal(err)
	}
	if !c.Equals(c2) {
		t.Errorf("cids should be the same: %s %s", c, c2)
	}
}
