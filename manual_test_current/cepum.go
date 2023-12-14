package main

import (
	"log"
	"time"

	"github.com/1f604/util"
)

func main() {

	log.SetFlags(log.Llongfile)

	cepum_params := util.CEPUMParams{
		Expiry_check_interval_seconds_ram:  1,
		Expiry_check_interval_seconds_disk: 1,
		Extra_keeparound_seconds_ram:       1,
		Extra_keeparound_seconds_disk:      30,
		Bucket_interval:                    20,
		Bucket_directory_path_absolute:     "/tmp/buckets",
		Size_file_path_absolute:            "/tmp/size/size.txt",
		B53m:                               util.NewBase53IDManager(),
		Size_file_rounded_multiple:         8,
		Generate_strings_up_to:             2,
	}

	cepum := util.CreateConcurrentExpiringPersistentURLMapFromDisk(&cepum_params)
	go func() {
		for {
			time.Sleep(200 * time.Millisecond)
			expiry_timestamp := time.Now().Unix() + int64(150)
			cepum.PutEntry(3, "google.com", expiry_timestamp)
			//log.Println("Added new entry:", val, err)
			//log.Println("Current timestamp", time.Now().Unix())
		}
	}()

	for i := 0; i < 20; i++ {
		rand_num, _ := util.Crypto_Randint(90)

		expiry_timestamp := time.Now().Unix() + int64(rand_num)
		cepum.PutEntry(3, "google.com", expiry_timestamp)
		//log.Println(val, err)
		//log.Println("Current timestamp", time.Now().Unix())
	}

	for {
		time.Sleep(1 * time.Second)
		//cepum.PrintInternalState()
	}

	log.Println("wtf!!!!!!!!!!!ww!")
}
