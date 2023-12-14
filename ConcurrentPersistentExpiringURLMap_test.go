// These tests do not test thread safety. It is just assumed that the accessors are thread safe because they are guarded by a mutex.
// The tests are for basic functionality to check that adding, getting, and expiring entries give expected results without errors.

package util_test

import (
	"testing"

	"github.com/1f604/util"
)

func Test_CPEUM_AddRestartReload(t *testing.T) {
	t.Parallel()

	// In this test, we add some entries then restart the server then reload and see if it all works.
	cepum_params := util.CEPUMParams{}
	util.CreateConcurrentExpiringPersistentURLMapFromDisk(&cepum_params)
}
