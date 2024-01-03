package main

import (
	"fmt"
	"os"

	"github.com/1f604/util"
)

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) != 1 {
		fmt.Println("Usage: gen_random_id [length]")
		os.Exit(1)
	}
	length, err := util.String_to_int64(os.Args[1])
	util.Check_err(err)

	b53m := util.NewBase53IDManager()
	id, _ := b53m.B53_generate_random_Base53ID(int(length))

	fmt.Println(id)
}
