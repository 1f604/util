package main

import (
	"fmt"
	"time"

	"github.com/1f604/util"
)

func main() {

	b53m := util.NewBase53IDManager()

	cepum_args := util.CEPUMParams{
		Expiry_check_interval_seconds:   5,
		Extra_keeparound_seconds_ram:    60,
		Extra_keeparound_seconds_disk:   100,
		Bucket_directory_path_absolute:  "/tmp/buckets",
		Bucket_interval:                 30,
		B53m:                            b53m,
		Size_file_rounded_growth_amount: 1000,
		Create_size_file_if_not_exists:  true,
		Generate_strings_up_to:          3,
	}

	cepum := util.CreateConcurrentExpiringPersistentURLMapFromDisk(&cepum_args)

	str, err := cepum.GetEntry("00")
	fmt.Println("str:", str)
	fmt.Println("err:", err)

	curtime_stamp := time.Now().Unix()
	str, err = cepum.PutEntry(2, "google.com", curtime_stamp+5)
	fmt.Println("str:", str)
	fmt.Println("err:", err)
}
