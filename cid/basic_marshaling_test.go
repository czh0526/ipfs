package cid

import (
	"bytes"
	"testing"

	gcid "github.com/ipfs/go-cid"
	mbase "github.com/multiformats/go-multibase"
	mh "github.com/multiformats/go-multihash"
)

func assertEqual(t *testing.T, a, b gcid.Cid) {
	if a.Type() != b.Type() {
		t.Fatal("mismatch on type")
	}

	if a.Version() != b.Version() {
		t.Fatal("mismatch on version")
	}

	if !bytes.Equal(a.Hash(), b.Hash()) {
		t.Fatal("multihash mismatch")
	}
}

func TestBasicMarshaling(t *testing.T) {
	// 计算 Multihash
	h, err := mh.Sum([]byte("TEST"), mh.SHA3, 4)
	if err != nil {
		t.Fatalf("hash error: %s", err)
	}

	// 通过给 Multihash 指定 Base编码，构建 Cid
	cid := gcid.NewCidV1(mbase.Base8, h)
	// Cast() 测试 => 通过 Cast()构造 Cid 对象
	data := cid.Bytes()
	out, err := gcid.Cast(data)
	if err != nil {
		t.Fatalf("get data error: %s", err)
	}
	assertEqual(t, cid, out)

	// Decode() 测试 => 通过 Decode()构造 Cid 对象
	s := cid.String()
	out2, err := gcid.Decode(s)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, cid, out2)
}
