package main

import (
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
	"math/rand"
)

func main() {
	str := UuidBase58()
	fmt.Println(str)
}

func UuidBase58() string {
	var oid []byte
	orderId, err := uuid.NewUUID()
	if err != nil {
		oid = RandByte(16)
	} else {
		oid, _ = orderId.MarshalBinary()
	}
	return base58.Encode(oid)
}

func RandByte(num int) []byte {
	var key []byte
	for i := 0; i < num; i++ {
		key = append(key, uint8(rand.Intn(256)))
	}
	return key
}
