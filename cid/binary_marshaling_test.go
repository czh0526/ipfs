package cid

import (
	"testing"

	gcid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

func TestBinaryMarshaling(t *testing.T) {
	data := []byte("this is some test content")
	hash, _ := mh.Sum(data, mh.SHA2_256, -1)
	c := gcid.NewCidV1(gcid.DagCBOR, hash)
	var c2 gcid.Cid

	// 测试 Cid 的 MarshalBinary() 方法
	data, err := c.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	// 测试 Cid 的 UnmarshalBinary() 方法
	err = c2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	if !c.Equals(c2) {
		t.Errorf("cids should be the same: %s %s", c, c2)
	}
}
