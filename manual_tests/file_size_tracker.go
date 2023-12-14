package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/1f604/util"
)

func run(n int) {
	filename := "/tmp/testfile"
	os.Remove(filename)
	f, err := os.OpenFile("/tmp/testfile", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	total_size := 0
	for i := 0; i < 1000; i++ {
		rand_length, err := util.Crypto_Randint(n)
		util.Check_err(err)
		new_string := strings.Repeat("x", rand_length)
		n, err := f.Write([]byte(new_string))
		total_size += n

		// check file
		slice := make([]int64, 1220000)
		// how long does it take to run this function 5000 times
		start := time.Now()
		for j := 0; j < 1000000; j++ {
			filesize := util.Get_file_size(f)
			slice[j] = filesize
		}
		end := time.Now().Sub(start)
		var total int64 = 0
		for _, c := range slice {
			total += c
		}
		fmt.Println(total)
		fmt.Println("end:", end)

		filesize := util.Get_file_size(f)
		if filesize != int64(total_size) {
			fmt.Println(total_size, filesize)
			panic("Not equal")
		}
		//fmt.Println(total_size, filesize)
	}
}

func main() {
	// generate string of random length and write it to file
	for i := 1; i < 200; i++ {
		run(i * 5)
	}
	fmt.Println("all done.")
}
