package main

import (
	"crypto/sha1"
	"math/big"
)

const keySize = 128

var two = big.NewInt(2)
var hashMod = new(big.Int).Exp(big.NewInt(2), big.NewInt(keySize), nil)

func HashString(elt string) *big.Int {
	hasher := sha1.New()
	hasher.Write([]byte(elt))
	return new(big.Int).SetBytes(hasher.Sum(nil))
}

func GetMod(elt *big.Int) *big.Int {
	return new(big.Int).Mod(elt, hashMod)
}

func Jump(address string, fingerentry int) *big.Int {
	n := HashString(address)
	fingerentryminus1 := big.NewInt(int64(fingerentry))
	jump := new(big.Int).Exp(two, fingerentryminus1, nil)
	sum := new(big.Int).Add(n, jump)

	return new(big.Int).Mod(sum, hashMod)
}

func Between(start, elt, end *big.Int, inclusive bool) bool {
	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}
}
func ConvToBig(b []byte) *big.Int {
	return new(big.Int).SetBytes(b)
}
