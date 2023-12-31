// You can edit this code!
// Click here and start typing.
package main

import (
	"fmt"
	"math/big"
	"time"

	"github.com/1f604/util"
)

// Naive algorithm, only suitable for small b.
func Power_Naive(a, b int) int {
	multiplier := a
	for i := 1; i < b; i++ {
		a *= multiplier
	}
	return a
}

func Power_Slow(a, b, m int) int {
	result := new(big.Int).Exp(
		big.NewInt(int64(a)),
		big.NewInt(int64(b)),
		big.NewInt(int64(m)),
	)
	return int(result.Int64())
}

func main() {
	start := time.Now()
	fmt.Println(Power_Naive(2, 23))
	fmt.Println(time.Now().Sub(start))

	fmt.Println(util.Crypto_RandString(0))
	fmt.Println(util.Crypto_RandString(0))
	fmt.Println(util.Crypto_RandString(1))
	fmt.Println(util.Crypto_RandString(1))
	fmt.Println(util.Crypto_RandString(1))
	fmt.Println(util.Crypto_RandString(2))
	fmt.Println(util.Crypto_RandString(2))
	fmt.Println(util.Crypto_RandString(2))
	fmt.Println(util.Crypto_RandString(12))
	fmt.Println(util.Crypto_RandString(12))
}
