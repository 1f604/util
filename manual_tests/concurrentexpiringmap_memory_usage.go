package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/1f604/util"
)

func main() {

	util.PrintMemUsage()

	cur_time := time.Now().Unix() - 5

	value_string := strings.Repeat("#", 128)

	begin := time.Now()
	var cem_list []util.CEMItem
	for j := 10; j < 20; j++ {
		for i := 0; i < 1000000; i++ {
			keystr := util.Int64_to_string(int64(j)) + "$" + util.Int64_to_string(int64(i))
			cem_list = append(cem_list, util.CEMItem{
				Key:              keystr,
				Value:            keystr + value_string,
				Expiry_time_unix: cur_time,
			})
			//cem.Put_New_Entry(keystr, keystr+value_string, cur_time)
		}
		// Force GC to clear up, should see a memory drop
		runtime.GC()
	}
	fmt.Printf("Bulk generate took %q seconds\n", time.Now().Sub(begin))

	begin = time.Now()
	cem := util.NewConcurrentExpiringMapFromSlice(cem_list)
	fmt.Printf("Bulk add took %q seconds\n", time.Now().Sub(begin))

	util.PrintMemUsage()

	fmt.Println("Removing all expired...")
	begin = time.Now()
	cem.Remove_All_Expired(0)
	fmt.Printf("Removing took %q seconds\n", time.Now().Sub(begin))
	fmt.Println("Done.")
	util.PrintMemUsage()

	begin = time.Now()
	for j := 20; j < 30; j++ {
		for i := 0; i < 1000000; i++ {
			keystr := util.Int64_to_string(int64(j)) + "$" + util.Int64_to_string(int64(i))
			cem.Put_New_Entry(keystr, keystr+value_string, cur_time)
		}
		// Force GC to clear up, should see a memory drop
		runtime.GC()
	}
	fmt.Printf("Individual add took %q seconds\n", time.Now().Sub(begin))

	util.PrintMemUsage()
}
