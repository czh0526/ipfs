package cid

import (
	"fmt"
	"strings"
	"testing"

	gcid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

func TestParse(t *testing.T) {
	cid, err := gcid.Parse(123)
	if err == nil {
		t.Fatalf("expected error from Parse()")
	}

	if !strings.Contains(err.Error(), "can't parse 123 as Cid") {
		t.Fatalf("expected int error, got %s", err.Error())
	}

	theHash := "QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n"
	h, err := mh.FromB58String(theHash)
	if err != nil {
		t.Fatalf("mh.FromB58String() error: %s", err)
	}

	assertions := [][]interface{}{
		// cidv0
		[]interface{}{gcid.NewCidV0(h), theHash},
		// cidv0 bytes 形式
		[]interface{}{gcid.NewCidV0(h).Bytes(), theHash},
		// Multihash 形式
		[]interface{}{h, theHash},
		// value 形式
		[]interface{}{theHash, theHash},
		// ipfs/<str> 形式
		[]interface{}{"/ipfs/" + theHash, theHash},
		[]interface{}{"https://ipfs.io/ipfs/" + theHash, theHash},
		[]interface{}{"http://localhost:8080/ipfs/" + theHash, theHash},
	}

	assert := func(arg interface{}, expected string) error {
		cid, err = gcid.Parse(arg)
		if err != nil {
			return err
		}

		if cid.Version() != 0 {
			return fmt.Errorf("expectd version 0, got %s", string(cid.Version()))
		}
		actual := cid.Hash().B58String()
		if actual != expected {
			return fmt.Errorf("expected hash %s, got %s", expected, actual)
		}
		actual = cid.String()
		if actual != expected {
			return fmt.Errorf("expected string %s, got %s", expected, actual)
		}
		return nil
	}

	for _, args := range assertions {
		err := assert(args[0], args[1].(string))
		if err != nil {
			t.Fatalf("assert error: %s", err)
		}
		fmt.Printf("assert good ==> %v \n", args[0])
	}
}
