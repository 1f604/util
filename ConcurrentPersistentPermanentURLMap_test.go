// These tests do not test thread safety. It is just assumed that the accessors are thread safe because they are guarded by a mutex.
// The tests are for basic functionality to check that adding, getting, and expiring entries give expected results without errors.

package util_test

import (
	"log"
	"testing"

	"github.com/1f604/util"
)

func Test_CPPUM_AddRestartReload(t *testing.T) {
	t.Parallel()

	log.SetFlags(log.Llongfile) // tell the logger to only log the file name where the log.print function is called, we'll add in the date manually.

	cppum_params := util.CPPUMParams{
		Log_directory_path_absolute: "/tmp/logs/",
		B53m:                        util.NewBase53IDManager(),
		Generate_strings_up_to:      3,
		Log_file_max_size_bytes:     300,
		Size_file_rounded_multiple:  5,
		Size_file_path_absolute:     "/tmp/sizefile/size.txt",
	}

	cppum := util.CreateConcurrentPersistentPermanentURLMapFromDisk(&cppum_params)

	val, err := cppum.PutEntry(2, "google.com", 0, util.TYPE_MAP_ITEM_URL)
	log.Println(val, err)
}
