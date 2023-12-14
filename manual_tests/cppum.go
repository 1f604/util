package main

import (
	"log"

	"github.com/1f604/util"
)

func main() {

	log.SetFlags(log.Llongfile)

	cppum_params := util.CPPUMParams{
		Log_directory_path_absolute: "/tmp/logs/",
		B53m:                        util.NewBase53IDManager(),
		Generate_strings_up_to:      2,
		Log_file_max_size_bytes:     300,
		Size_file_rounded_multiple:  8,
		Size_file_path_absolute:     "/tmp/sizefile/size.txt",
	}

	cppum := util.CreateConcurrentPersistentPermanentURLMapFromDisk(&cppum_params)

	for i := 0; i < 3; i++ {
		val, err := cppum.PutEntry(3, "google.com")
		log.Println(val, err)
	}

	cppum.PrintInternalState()
}
