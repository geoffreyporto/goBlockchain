package main

import (
	"reflect"
	"testing"
)

func TestInit(t *testing.T) {
	b = new(Blockchain)
	b.Init()

	block := b.GetLatestBlock()
	if len(b.blockchain) != 1 {
		t.Error("Init failed")
	}
	if block.Data != "Gensis Block" {
		t.Error("Genesis block incorrectly generated")
	}
}

func TestCalculateHash(t *testing.T) {
	var v []byte
	e := make([]byte, 32)
	v = CalculateHash(1, -1, e, "A")
	if len(v) != 32 {
		t.Error("Unexpected hash length")
	}
}

func TestCalculateHashForBlock(t *testing.T) {
	e := make([]byte, 32)
	f := make([]byte, 32)
	f[1] = 1

	block1 := &(Block{0, 0, e, "A", e})
	block2 := &(Block{0, 0, e, "A", f})
	block3 := &(Block{0, 1, e, "A", e})

	hash1 := CalculateHashForBlock(block1)
	hash2 := CalculateHashForBlock(block2)
	hash3 := CalculateHashForBlock(block3)

	if !reflect.DeepEqual(hash1, hash2) {
		t.Error("hash for same block not matching")
	}
	if reflect.DeepEqual(hash1, hash3) {
		t.Error("hashes for different blocks are matching")
	}
}

func TestGenerateNextBlock(t *testing.T) {
	b = new(Blockchain)
	b.Init()

	b.GenerateNextBlock("BlockNext")
	if len(b.blockchain) != 2 {
		t.Error("Second block not generated")
	}

	if b.blockchain[1].Data != "BlockNext" {
		t.Error("Second block data incorrect")
	}
}

func TestIsValidNewBlock(t *testing.T) {
	b = new(Blockchain)
	b.Init()

	b.GenerateNextBlock("Next")
	block1 := &b.blockchain[0]
	block2 := &b.blockchain[1]
	block3 := *block2
	// change hashes to be incorrect
	block3.PrevHash[0] = 0
	block3.Hash[0] = 0

	if IsValidNewBlock(block1, block1) {
		t.Error("Same index block is accepted as new block")
	}
	if IsValidNewBlock(block1, &block3) {
		t.Error("Incorrect hashes was accepted as valid")
	}
}
