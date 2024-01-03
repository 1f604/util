// Create 200k files in one directory vs 200k files in 100 separate directories
// See if speed of accessing files is affected
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"math/rand"

	"github.com/1f604/util"
)

func main() {
	// Now time read and write times
	fmt.Println("Now creating 100 x 10kb files in a small directory.")
	start := time.Now()
	for j := 0; j < 100; j++ {
		num1 := rand.Intn(100)
		num2 := rand.Intn(2000)
		f, err := os.Create("/tmp/dir" + util.Int64_to_string(int64(num1)) + "/test" + util.Int64_to_string(int64(num2)) + ".txt")
		if err != nil {
			log.Fatal(err)
		}

		if err := f.Truncate(1e5); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Time taken:", time.Now().Sub(start))

	fmt.Println("Now reading 1000 random 10kb files in a small directory.")
	start = time.Now()
	list := [][]byte{}
	for j := 0; j < 1000; j++ {
		num1 := rand.Intn(100)
		num2 := rand.Intn(2000)
		contents, err := os.ReadFile("/tmp/dir" + util.Int64_to_string(int64(num1)) + "/" + util.Int64_to_string(int64(num2)) + ".txt")
		if err != nil {
			log.Fatal(err)
		}
		list = append(list, contents)
	}
	fmt.Println("Time taken:", time.Now().Sub(start))

	fmt.Println("Now creating 100 x 10kb files in a big directory.")
	start = time.Now()
	for j := 0; j < 100; j++ {
		f, err := os.Create("/tmp/bigdir/test" + util.Int64_to_string(int64(j)) + ".txt")
		if err != nil {
			log.Fatal(err)
		}

		if err := f.Truncate(1e5); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Time taken:", time.Now().Sub(start))

	fmt.Println("Now reading 1000 random 10kb files in a big directory.")
	start = time.Now()
	for j := 0; j < 1000; j++ {
		num := rand.Intn(200000)
		contents, err := os.ReadFile("/tmp/bigdir/" + util.Int64_to_string(int64(num)) + ".txt")
		if err != nil {
			log.Fatal(err)
		}
		list = append(list, contents)
	}
	fmt.Println("Time taken:", time.Now().Sub(start))
}
