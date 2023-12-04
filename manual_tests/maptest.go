package main

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/1f604/util"
)

type KVPair = util.CEMItem

/*
func add_to_map(testmap sync.Map, pairs []KVPair, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, pair := range pairs {
		testmap.Store(pair.Key, pair.Value)
		//[pair.key] = pair.value
	}
}
*/

var mut sync.Mutex

func main() {
	///
	//testmap := make(map[string]string, 8000000)
	//var testmap sync.Map

	pairs := make([]KVPair, 0)

	value_string := strings.Repeat("#", 128)

	for j := 10; j < 20; j++ {
		for i := 0; i < 1000000; i++ {
			keystr := util.Int64_to_string(int64(j)) + "$" + util.Int64_to_string(int64(i))
			valstr := keystr + "$$" + value_string
			pairs = append(pairs, KVPair{
				Key:              keystr,
				Value:            valstr,
				Expiry_time_unix: time.Now().Unix(),
			})
		}
	}
	// Force GC to clear up, should see a memory drop
	runtime.GC()

	start := time.Now()
	/*
		var wg sync.WaitGroup
		for j := 0; j < 10; j++ {
			wg.Add(1)
			go add_to_map(testmap, pairs[j*1000000:(j+1)*1000000], &wg)
		}
		wg.Wait()
	*/
	util.NewConcurrentExpiringMapFromSlice(pairs)
	fmt.Println("Time elapsed:", time.Now().Sub(start))

}
