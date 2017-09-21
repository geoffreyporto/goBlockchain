package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCalculateHash(t *testing.T) {
	var v []byte
	e := make([]byte, 32)
	v = CalculateHash(1, -1, e, e)
	if len(v) != 32 {
		t.Error("Unexpected hash length")
	}
}

func TestMarshalling(t *testing.T) {
	var block = *GetGenesisBlock()
	blockMarshalled, err := json.Marshal(block)
	if err != nil {
		t.Error(err)
	}

	var block1 Block
	err1 := json.Unmarshal(blockMarshalled, &block1)

	if err1 != nil {
		t.Error(err1)
	}

	if !reflect.DeepEqual(block, block1) {
		t.Error("Marshalling and Unmarshalling failed")
	}
}
