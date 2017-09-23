package main

import (
	"crypto/sha256"
	"encoding/binary"
	"log"
<<<<<<< HEAD
	"reflect"
=======
>>>>>>> 167e2284bd51e29b581146319b5ab391bb7cca1a
	"time"
)

// Block is the building blocks of a block chain
type Block struct {
	Index     uint64 `json:"id"`
	Timestamp int64  `json:"timestamp"`
	PrevHash  []byte `json:"prevhash"`
	Data      []byte `json:"data"`
	Hash      []byte `json:"hash"`
}

// Blockchain stores the blockchain
type Blockchain struct {
	blockchain []Block
}

const (
	HashLen = 32
	IntSize = 8
)

// Init initializes blockchain struct by creating gensis block
func Init(b *Blockchain) {
	// Generate genesis block
	var e = make([]byte, HashLen)
	timeNow := time.Now().UTC().Unix()
	genesisHash := CalculateHash(0, timeNow, e, e)
	genesisBlock := Block{0, timeNow, e, e, genesisHash}

	// Append to blockchain
	b.blockchain = append(b.blockchain, genesisBlock)
}

// CalculateHash calculates the hash of a block
func CalculateHash(index uint64, ts int64, prevHash []byte, data []byte) []byte {
	if len(prevHash) != HashLen {
		log.Fatal("CalculateHash(): Previous hash is not 32 bytes")
	}

	h := sha256.New()

	b := make([]byte, IntSize)
	binary.LittleEndian.PutUint64(b, index)
	h.Write(b)

	// uint64 conversion doesn't change the sign bit, only the way it's interpreted
	t := make([]byte, IntSize)
	binary.LittleEndian.PutUint64(t, uint64(ts))
	h.Write(t)

	h.Write(prevHash)
	h.Write(data)

	log.Printf("Calculated hash: %x\n", h.Sum(nil))
	return h.Sum(nil)
}

// CalculateHashForBlock calculates the has for the block
func CalculateHashForBlock(block *Block) []byte {
	return CalculateHash(block.Index, block.Timestamp, block.PrevHash, block.Data)
}

// GenerateNextBlock generates next block in the chain
func GenerateNextBlock(b *Blockchain, blockData []byte) *Block {
	prevBlock := b.GetLatestBlock()
	nextIndex := prevBlock.Index + 1
	nextTimestamp := time.Now().UTC().Unix()
	nextHash := CalculateHash(nextIndex, nextTimestamp, prevBlock.PrevHash, blockData)

	return &Block{nextIndex, nextTimestamp, prevBlock.PrevHash, blockData, nextHash}
}

// GetGenesisBlock retrives the first block in the chain
func GetGenesisBlock(b *Blockchain) *Block {
	if len(b.blockchain) < 1 {
		log.Println("Did not initialize blockchain")
	}
	return &b.blockchain[0]
}

// GetLatestBlock retrieves the last block in the chain
func (b *Blockchain) GetLatestBlock() *Block {
	// returns last element in slice
	return &b.blockchain[len(b.blockchain)-1]
}

// IsValidNewBlock checks the integrity of the newest block
func IsValidNewBlock(newBlock, prevBlock *Block) bool {
	if prevBlock.Index+1 != newBlock.Index {
		log.Println("IsValidNewBlock(): invalid index")
		return false
	} else if !reflect.DeepEqual(prevBlock.Hash, newBlock.PrevHash) {
		log.Println("IsValidNewBlock(): invalid previous hash")
		return false
	} else if !reflect.DeepEqual(CalculateHashForBlock(newBlock), newBlock.Hash) {
		log.Printf("invalid hash: %x is not %x\n", CalculateHashForBlock(newBlock), newBlock.Hash)
		return false
	}
	return true
}

// IsValidChain checks if the chain received is valid
func IsValidChain(newBlockchain []Block) bool { // TODO: pass by ref?
	// newBlockchain is in JSON format
	// Check if the genesis block matches
	if !reflect.DeepEqual(newBlockchain[0], blockchain[0]) {
		return false
	}

	var tempBlocks []Block
	tempBlocks = append(tempBlocks, newBlockchain[0])
	for i := 1; i < len(newBlockchain); i++ {
		if IsValidNewBlock(&newBlockchain[i], &tempBlocks[i-1]) {
			tempBlocks = append(tempBlocks, newBlockchain[i])
		} else {
			return false
		}
	}

	// All blocks are valid in the new blockchain
	return true
}
