// These tests do not test thread safety. It is just assumed that the accessors are thread safe because they are guarded by a mutex.
// The tests are for basic functionality to check that adding, getting, and expiring entries give expected results without errors.

package util_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/1f604/util"
)

func Test_ConcurrentExpiringMap(t *testing.T) {
	t.Parallel()

	var err error
	cur_time := time.Now().Unix()

	// Some items and their priorities.
	items := map[string]int64{
		"banana":  cur_time - 1,
		"apple":   cur_time - 2,
		"pear":    cur_time + 1,
		"peaches": cur_time + 2,
	}

	cem := util.NewEmptyConcurrentExpiringMap(nil)
	for key, expiry_time := range items {
		err = cem.Put_New_Entry(key, expiry_time, expiry_time)
		util.Assert_no_error(t, err, 1)
	}

	// Try to insert something that has already been inserted
	err = cem.Put_New_Entry("banana", "key", cur_time+9999)
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: key already exists", 1)

	// Try to fetch something that hasn't been expired
	value, err := cem.Get_Entry("peaches")
	util.Assert_result_equals_interface(t, value, err, cur_time+2, 1)
	value, err = cem.Get_Entry("pear")
	util.Assert_result_equals_interface(t, value, err, cur_time+1, 1)

	// Try to fetch something that doesn't exist
	_, err = cem.Get_Entry("ballast")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)

	// Try to fetch something that has expired
	_, err = cem.Get_Entry("banana")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: key expired", 1)
	_, err = cem.Get_Entry("apple")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: key expired", 1)

	// Remove expired items
	cem.Remove_All_Expired(0)

	// Try to fetch items again
	value, err = cem.Get_Entry("peaches")
	util.Assert_result_equals_interface(t, value, err, cur_time+2, 1)
	value, err = cem.Get_Entry("pear")
	util.Assert_result_equals_interface(t, value, err, cur_time+1, 1)
	_, err = cem.Get_Entry("ballast")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)
	_, err = cem.Get_Entry("banana")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)
	_, err = cem.Get_Entry("apple")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)

	// Insert new items
	_ = cem.Put_New_Entry("oranges", cur_time-1, cur_time-1)
	_ = cem.Put_New_Entry("squares", cur_time-2, cur_time-2)
	_ = cem.Put_New_Entry("jeremys", cur_time-5, cur_time-5)

	// Remove expired items
	cem.Remove_All_Expired(3)

	// Try to fetch items again
	value, err = cem.Get_Entry("peaches")
	util.Assert_result_equals_interface(t, value, err, cur_time+2, 1)
	value, err = cem.Get_Entry("pear")
	util.Assert_result_equals_interface(t, value, err, cur_time+1, 1)
	_, err = cem.Get_Entry("ballast")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)
	_, err = cem.Get_Entry("banana")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)
	_, err = cem.Get_Entry("apple")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)
	_, err = cem.Get_Entry("oranges")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: key expired", 1)
	_, err = cem.Get_Entry("squares")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: key expired", 1)
	_, err = cem.Get_Entry("jeremys")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)
}

func Test_ConcurrentExpiringMap_LoadEntriesBulk(t *testing.T) {
	t.Parallel()

	var err error
	cur_time := time.Now().Unix()

	// Some items and their priorities.
	items := map[string]int64{
		"banana":  cur_time - 1,
		"apple":   cur_time - 2,
		"pear":    cur_time + 1,
		"peaches": cur_time + 2,
		"oranges": cur_time - 2,
	}
	cem_list := make([]util.CEMItem, 0)

	for key, expiry_time := range items {
		cem_list = append(cem_list, util.CEMItem{
			Key:              key,
			Value:            expiry_time,
			Expiry_time_unix: expiry_time,
		})
	}

	cem := util.NewConcurrentExpiringMapFromSlice(nil, cem_list)

	// Try to fetch items again
	value, err := cem.Get_Entry("peaches")
	util.Assert_result_equals_interface(t, value, err, cur_time+2, 1)
	value, err = cem.Get_Entry("pear")
	util.Assert_result_equals_interface(t, value, err, cur_time+1, 1)
	_, err = cem.Get_Entry("ballast")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)
	_, err = cem.Get_Entry("banana")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: key expired", 1)
	_, err = cem.Get_Entry("apple")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: key expired", 1)
	_, err = cem.Get_Entry("oranges")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: key expired", 1)
}

func Test_ConcurrentExpiringMap_Expiry_Callback(t *testing.T) {
	t.Parallel()

	var err error
	cur_time := time.Now().Unix()

	outer := []string{}

	expiry_callback := func(item interface{}) {
		outer = append(outer, item.(string))
	}
	fmt.Println("outer begin:", outer)

	// Some items and their priorities.
	items := map[string]int64{
		"banana":  cur_time - 1,
		"apple":   cur_time - 2,
		"pear":    cur_time + 1,
		"peaches": cur_time + 2,
	}

	cem := util.NewEmptyConcurrentExpiringMap(expiry_callback)
	for key, expiry_time := range items {
		err = cem.Put_New_Entry(key, expiry_time, expiry_time)
		util.Assert_no_error(t, err, 1)
	}

	// Try to insert something that has already been inserted
	err = cem.Put_New_Entry("banana", "key", cur_time+9999)
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: key already exists", 1)

	// Try to fetch something that hasn't been expired
	value, err := cem.Get_Entry("peaches")
	util.Assert_result_equals_interface(t, value, err, cur_time+2, 1)
	value, err = cem.Get_Entry("pear")
	util.Assert_result_equals_interface(t, value, err, cur_time+1, 1)

	// Try to fetch something that doesn't exist
	_, err = cem.Get_Entry("ballast")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)

	// Try to fetch something that has expired
	_, err = cem.Get_Entry("banana")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: key expired", 1)
	_, err = cem.Get_Entry("apple")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: key expired", 1)

	// should have no expired items
	util.Assert_result_equals_interface(t, len(outer), nil, 0, 1)
	// Remove expired items
	cem.Remove_All_Expired(0)
	// banana and apple should be expired now
	sort.Strings(outer)
	util.Assert_result_equals_string_slice(t, outer, nil, []string{"apple", "banana"}, 1)
	// fmt.Println("outer after expiry:", outer)

	// Try to fetch items again
	value, err = cem.Get_Entry("peaches")
	util.Assert_result_equals_interface(t, value, err, cur_time+2, 1)
	value, err = cem.Get_Entry("pear")
	util.Assert_result_equals_interface(t, value, err, cur_time+1, 1)
	_, err = cem.Get_Entry("ballast")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)
	_, err = cem.Get_Entry("banana")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)
	_, err = cem.Get_Entry("apple")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)

	// Insert new items
	_ = cem.Put_New_Entry("oranges", cur_time-1, cur_time-1)
	_ = cem.Put_New_Entry("squares", cur_time-2, cur_time-2)
	_ = cem.Put_New_Entry("jeremys", cur_time-5, cur_time-5)

	// Remove expired items
	fmt.Println("outer before expiry2:", outer)
	util.Assert_result_equals_string_slice(t, outer, nil, []string{"apple", "banana"}, 1)
	cem.Remove_All_Expired(3)
	sort.Strings(outer)
	util.Assert_result_equals_string_slice(t, outer, nil, []string{"apple", "banana", "jeremys"}, 1)
	fmt.Println("outer after expiry2:", outer)

	// Try to fetch items again
	value, err = cem.Get_Entry("peaches")
	util.Assert_result_equals_interface(t, value, err, cur_time+2, 1)
	value, err = cem.Get_Entry("pear")
	util.Assert_result_equals_interface(t, value, err, cur_time+1, 1)
	_, err = cem.Get_Entry("ballast")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)
	_, err = cem.Get_Entry("banana")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)
	_, err = cem.Get_Entry("apple")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)
	_, err = cem.Get_Entry("oranges")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: key expired", 1)
	_, err = cem.Get_Entry("squares")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: key expired", 1)
	_, err = cem.Get_Entry("jeremys")
	util.Assert_error_equals(t, err, "ConcurrentExpiringMap: nonexistent key", 1)

}
