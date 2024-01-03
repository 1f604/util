// Create 200k files in one directory vs 200k files in 100 separate directories
// See if speed of accessing files is affected
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/1f604/util"
)

func main() {
	// First, create 100 directories
	filepaths := []string{}
	for i := 0; i < 100; i++ {
		newfilepath := "/tmp/dir" + util.Int64_to_string(int64(i)) + "/"
		filepaths = append(filepaths, newfilepath)
		err := os.MkdirAll(newfilepath, os.ModePerm)
		util.Check_err(err)
	}
	fmt.Println("Created 100 directories.")
	// Next, create 2k files in each directory
	fmt.Println("Now creating 2k x 10kb files in each small directory.")
	for i := 0; i < 100; i++ {
		for j := 0; j < 2000; j++ {
			f, err := os.Create("/tmp/dir" + util.Int64_to_string(int64(i)) + "/" + util.Int64_to_string(int64(j)) + ".txt")
			if err != nil {
				log.Fatal(err)
			}

			if err := f.Truncate(1e5); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Next, create 200k files in one directory
	fmt.Println("Now creating 200k x 10kb files in one big directory.")
	for j := 0; j < 200000; j++ {
		f, err := os.Create("/tmp/bigdir/" + util.Int64_to_string(int64(j)) + ".txt")
		if err != nil {
			log.Fatal(err)
		}

		if err := f.Truncate(1e5); err != nil {
			log.Fatal(err)
		}
	}

	// Now time read and write times
	fmt.Println("Now creating 100 x 10kb files in a small directory.")
	start := time.Now()
	for j := 0; j < 100; j++ {
		f, err := os.Create("/tmp/dir1/test" + util.Int64_to_string(int64(j)) + ".txt")
		if err != nil {
			log.Fatal(err)
		}

		if err := f.Truncate(1e5); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Time taken:", time.Now().Sub(start))

	fmt.Println("Now reading 100 random 10kb files in a small directory.")
	start = time.Now()
	list := [][]byte{}
	for j := 0; j < 100; j++ {
		num, err := util.Crypto_Randint(2000)
		util.Check_err(err)
		contents, err := os.ReadFile("/tmp/dir2/" + util.Int64_to_string(int64(num)) + ".txt")
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

	fmt.Println("Now reading 100 random 10kb files in a big directory.")
	start = time.Now()
	for j := 0; j < 100; j++ {
		num, err := util.Crypto_Randint(200000)
		util.Check_err(err)
		contents, err := os.ReadFile("/tmp/bigdir/" + util.Int64_to_string(int64(num)) + ".txt")
		if err != nil {
			log.Fatal(err)
		}
		list = append(list, contents)
	}
	fmt.Println("Time taken:", time.Now().Sub(start))
}
