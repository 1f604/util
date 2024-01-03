package main

import (
	"fmt"
	"time"

	"github.com/1f604/util"
)

func main() {
	ebs := util.NewExpiringBucketStorage(5, "/tmp/buckets/", 5)
	s := ebs.InsertFile([]byte("Hello world!"), 1804234080)

	go util.RunFuncEveryXSeconds(ebs.DeleteExpiredBuckets, 6)
	fmt.Println(s)
	time.Sleep(20 * time.Second)
}
