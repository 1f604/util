// These tests do not test thread safety. It is just assumed that the accessors are thread safe because they are guarded by a mutex.
// The tests are for basic functionality to check that adding, getting, and expiring entries give expected results without errors.

package util_test

import (
	"testing"

	"github.com/1f604/util"
)

func Test_ConcurrentPermanentMap(t *testing.T) {
	t.Parallel()

	cpm := util.NewEmptyConcurrentPermanentMap()
	cpm.Put_New_Entry("key!", "value!", 0, util.TYPE_MAP_ITEM_URL)

	_, err := cpm.Get_Entry("key")
	util.Assert_error_equals(t, err, "ConcurrentPermanentMap: nonexistent key", 1)

	val, err := cpm.Get_Entry("key!")
	util.Assert_no_error(t, err, 1)
	util.Assert_result_equals_interface(t, val.GetValue(), nil, "value!", 1)
}
