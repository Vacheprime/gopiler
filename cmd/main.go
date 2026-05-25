package main

import (
	"fmt"

	"github.com/Vacheprime/gopiler"
)

func main() {
	bitSet := gopiler.NewBitSet(1)
	bitSet.Set(1)
	bitSet.Set(0)
	bitSet.Unset(1)
	fmt.Println(bitSet.Bits[0])
}
