package util_test

import (
	"fmt"
	"testing"

	"github.com/1f604/util"
)

func Test_GenericMapWithPasteCounter(t *testing.T) {
	t.Parallel()

	mwpc := util.NewMapWithPastesCount[util.MapItem](5)
	mwpc.InsertNew("hi", &util.ExpiringMapItem{})
	fmt.Println(mwpc.NumPastes())
}
