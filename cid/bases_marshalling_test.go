package cid

import (
	"testing"

	gcid "github.com/ipfs/go-cid"
	mbase "github.com/multiformats/go-multibase"
	mh "github.com/multiformats/go-multihash"
)

func TestBasesMarshalling(t *testing.T) {
	// 计算 hash
	h, err := mh.Sum([]byte("TEST"), mh.SHA3, 4)
	if err != nil {
		t.Fatal(err)
	}

	// 构建 cid
	cid := gcid.NewCidV1(mbase.Base8, h)

	testBases := []mbase.Encoding{
		mbase.Base16,
		mbase.Base32,
		mbase.Base32hex,
		mbase.Base32pad,
		mbase.Base32hexPad,
		mbase.Base58BTC,
		mbase.Base58Flickr,
		mbase.Base64pad,
		mbase.Base64urlPad,
		mbase.Base64url,
		mbase.Base64,
	}

	for _, b := range testBases {
		// StringOfBase() 测试
		s, err := cid.StringOfBase(b)
		if err != nil {
			t.Fatal(err)
		}

		if s[0] != byte(b) {
			t.Fatal("Invalid multibase header")
		}

		// 测试 Base 编码是否成功
		out2, err := gcid.Decode(s)
		if err != nil {
			t.Fatal(err)
		}

		assertEqual(t, cid, out2)
	}
}
