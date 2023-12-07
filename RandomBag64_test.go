package util_test

import (
	"slices"
	"testing"

	"github.com/1f604/util"
)

func Test_RandomBagDecreases(t *testing.T) {
	t.Parallel()

	items := []uint64{1, 2, 3, 4, 5}
	bag := util.CreateRandomBagFromSlice(items)

	util.Assert_result_equals_interface(t, bag.Size(), nil, 5, 1)

	_, err := bag.PopRandom()
	util.Check_err(err)

	util.Assert_result_equals_interface(t, bag.Size(), nil, 4, 1)
}

func Test_RandomBag_Returns_All_Items(t *testing.T) {
	t.Parallel()

	seen := make([]uint64, 500)
	for i := 0; i < 500; i++ {
		items := []uint64{1, 2, 3, 4, 5}
		bag := util.CreateRandomBagFromSlice(items)

		item, err := bag.PopRandom()
		util.Check_err(err)

		seen[i] = item
	}
	for i := 1; i < 6; i++ {
		if !slices.Contains(seen, uint64(i)) {
			panic("seen does not contain " + util.Int64_to_string(int64(i)))
		}
	}
}
