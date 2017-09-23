package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMarshalling(t *testing.T) {
	b := new(Blockchain)
	var block = *b.GetGenesisBlock()
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
