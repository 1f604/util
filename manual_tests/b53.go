package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/1f604/util"
)

func test_B53_generate_all_Base53IDs_int64_test(n int) ([]uint64, error) {
	result := make([]uint64, util.Power_Naive(53, n-1))
	for i := 0; i < len(result); i++ {
		result[i] = uint64(i)
	}

	return result, nil
}

func test2(n int, result *[]uint64) error {
	for i := 0; i < len(*result); i++ {
		(*result)[i] = uint64(i)
	}

	return nil
}

func timeit(b53m *util.Base53IDManager, n int) {
	util.PrintMemUsage()
	start := time.Now()
	// generate all length 2 ids
	results, _ := b53m.B53_generate_all_Base53IDs_int64_optimized(n)
	/*results := make([]uint64, util.Power(53, n-1, 0)+10)
	power := util.Power(53, n-1, 0)
	for i := 0; i < power; i++ {
		results[i] = 0
	}*/

	//test2(n, &results)

	//results, _ := test_B53_generate_all_Base53IDs_int64_test(n)
	/*results := make([]uint64, util.Power(53, n-1, 0))
	for i := 0; i < len(results); i++ {
		results[i] = uint64(i)
	}*/
	fmt.Println("n:", n, "seconds:", time.Now().Sub(start), "len:", len(results))
	debug.FreeOSMemory()
	runtime.GC()
	runtime.GC()
	debug.FreeOSMemory()
	runtime.GC()
	util.PrintMemUsage()
}

func main() {
	b53m := util.NewBase53IDManager()
	timeit(b53m, 2)
	timeit(b53m, 3)
	timeit(b53m, 4)
	timeit(b53m, 5)
	timeit(b53m, 5)
	timeit(b53m, 5)
	timeit(b53m, 5)
}
