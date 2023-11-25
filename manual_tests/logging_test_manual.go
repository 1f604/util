// This is a manual test for the log rotation deletion functionality
// You are supposed to run this file and then run `watch -n0.1 ls -alrt /tmp/logfiletest/` in a separate terminal to see that it's doing what it's supposed to do.
package main

import (
	"log"
	"os"
	"time"

	logging "github.com/1f604/util/logging"
)

func main() {

	logging_dir := "/tmp/logfiletest/"
	log_filename := "current.log"

	// First create test directory
	os.RemoveAll(logging_dir)
	os.MkdirAll(logging_dir, 0o755)

	// Now create a log file rotator

	logging.Set_up_logging_panic_on_err(logging_dir, log_filename, "log", 240, 1100)

	log.Print("hello")

	entries, err := os.ReadDir(logging_dir)
	if err != nil {
		log.Fatal(err)
	}

	actual := []string{}
	for _, e := range entries {
		actual = append(actual, e.Name())
	}

	for range time.Tick(time.Second * time.Duration(2)) {
		log.Print("testing.")
	}
}
