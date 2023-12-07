package util_test

import (
	"slices"
	"testing"

	"github.com/1f604/util"
)

func Test_Power_Naive(t *testing.T) {
	t.Parallel()

	base := 53
	for i := 0; i < 10; i++ {
		naive := util.Power_Naive(base, i)
		truth := util.Power_Slow(base, i, 0)
		util.Assert_result_equals_interface(t, naive, nil, truth, 1)
	}
}

func Test_Crypto_Randint(t *testing.T) {
	t.Parallel()

	slice := []int{}
	for i := 0; i < 500; i++ {
		x, err := util.Crypto_Randint(3)
		util.Check_err(err)
		slice = append(slice, x)
	}
	if !slices.Contains(slice, 0) {
		panic("slice does not contain 0")
	}
	if !slices.Contains(slice, 1) {
		panic("slice does not contain 1")
	}
	if !slices.Contains(slice, 2) {
		panic("slice does not contain 2")
	}
	if slices.Contains(slice, 3) {
		panic("slice contains 3")
	}
}
