package main

import (
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
