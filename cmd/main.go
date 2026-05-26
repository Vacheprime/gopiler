package main

import (
	"fmt"

	"github.com/Vacheprime/gopiler"
)

func main() {
	bitSet := gopiler.NewBitSet(10)
	bitSet.Set(1)
	bitSet.Set(0)
	bitSet.Set(54)
	for _, v := range bitSet.GetActiveBitPositions() {
		fmt.Println(v)
	}
}
