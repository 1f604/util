package main

import (
	"fmt"

	"github.com/1f604/util"
)

func main() {

	mwpc := util.NewMapWithPastesCount[util.MapItem](5)
	err := mwpc.InsertNew("hi2", util.NewTestExpiringMapItem("world", util.TYPE_MAP_ITEM_URL, 1))
	util.Check_err(err)
	fmt.Println(mwpc.NumPastes())
	_ = mwpc.InsertNew("hi2", util.NewTestExpiringMapItem("world", util.TYPE_MAP_ITEM_PASTE, 1))
	fmt.Println(mwpc.NumPastes())
	err = mwpc.InsertNew("hi", util.NewTestExpiringMapItem("world", util.TYPE_MAP_ITEM_PASTE, 1))
	util.Check_err(err)
	fmt.Println(mwpc.NumPastes())

	err = mwpc.InsertNew("hey", util.NewTestExpiringMapItem("world", util.TYPE_MAP_ITEM_URL, 1))
	util.Check_err(err)
	fmt.Println(mwpc.NumPastes())
	err = mwpc.InsertNew("hey2", util.NewTestExpiringMapItem("world", util.TYPE_MAP_ITEM_PASTE, 1))
	util.Check_err(err)
	fmt.Println(mwpc.NumPastes())

	fmt.Println("Now deleting...")
	// now test deletes
	mwpc.DeleteKey("hey")
	fmt.Println(mwpc.NumPastes())
	mwpc.DeleteKey("hi")
	fmt.Println(mwpc.NumPastes())
	mwpc.DeleteKey("hi")
	fmt.Println(mwpc.NumPastes())
	mwpc.DeleteKey("hey")
	fmt.Println(mwpc.NumPastes())
	mwpc.DeleteKey("hey2")
	fmt.Println(mwpc.NumPastes())
}
