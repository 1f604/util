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
	cpm.Put_Entry("key!", "value!")

	_, ok := cpm.Get_Entry("key")
	util.Assert_result_equals_bool(t, ok, nil, false, 1)

	val, ok := cpm.Get_Entry("key!")
	util.Assert_result_equals_bool(t, ok, nil, true, 1)
	util.Assert_result_equals_interface(t, val, nil, "value!", 1)

	// Some items and their priorities.
	items := map[interface{}]interface{}{
		"banana":  1,
		"apple":   2,
		"pear":    3,
		"peaches": 4,
	}

	cpm_list := make([]util.CPMItem, 0)

	for key, expiry_time := range items {
		cpm_list = append(cpm_list, util.CPMItem{
			Key:   key,
			Value: expiry_time,
		})
	}

	cpm = util.NewConcurrentPermanentMapFromSlice(cpm_list)

	_, ok = cpm.Get_Entry("key")
	util.Assert_result_equals_bool(t, ok, nil, false, 1)

	_, ok = cpm.Get_Entry("key!")
	util.Assert_result_equals_bool(t, ok, nil, false, 1)

	val, ok = cpm.Get_Entry("banana")
	util.Assert_result_equals_bool(t, ok, nil, true, 1)
	util.Assert_result_equals_interface(t, val, nil, 1, 1)

	val, ok = cpm.Get_Entry("apple")
	util.Assert_result_equals_bool(t, ok, nil, true, 1)
	util.Assert_result_equals_interface(t, val, nil, 2, 1)

	val, ok = cpm.Get_Entry("pear")
	util.Assert_result_equals_bool(t, ok, nil, true, 1)
	util.Assert_result_equals_interface(t, val, nil, 3, 1)

	val, ok = cpm.Get_Entry("peaches")
	util.Assert_result_equals_bool(t, ok, nil, true, 1)
	util.Assert_result_equals_interface(t, val, nil, 4, 1)
}
