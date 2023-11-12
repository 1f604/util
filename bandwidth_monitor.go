// Don't use this. It just counts the total bytes from your interface, which is not what you want (typically).
// Normally you want the total number of bytes THIS MONTH not ALL TIME, and this doesn't tell you that.
// And it can't do that. To do that it would have to know what the value was at the start of the month.
// To record that, it's probably best to use a separate program.

package util

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type BandwidthMonitor struct {
	total_all_bytes atomic.Int64
	total_tx_bytes  atomic.Int64
}

func (bm *BandwidthMonitor) RunThread(time_interval_secs int) {
	for range time.Tick(time.Second * time.Duration(time_interval_secs)) {
		update_stats(bm)
	}
}

func (bm *BandwidthMonitor) GetTotalAllBytes() int64 {
	return bm.total_all_bytes.Load()
}
func (bm *BandwidthMonitor) GetTotalTXBytes() int64 {
	return bm.total_tx_bytes.Load()
}

func read_number_from_file_panic_on_err(path string) int64 {
	file_bytes, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	// file_bytes contains newline at the end
	if file_bytes[len(file_bytes)-1] == 10 { //nolint:gomnd // 10 is newline
		file_bytes = file_bytes[:len(file_bytes)-1]
	}
	// convert to int64
	num, err := String_to_int64(string(file_bytes))
	if err != nil {
		panic(err)
	}
	return num
}

func update_stats(bm *BandwidthMonitor) {
	dir_entry_list, err := os.ReadDir("/sys/class/net")
	if err != nil {
		panic(err)
	}
	var sb strings.Builder
	var total_tx_bytes int64 = 0
	var total_all_bytes int64 = 0
	for _, direntry := range dir_entry_list {
		// For each dir entry, list the name and the tx bytes in a way that can be easily parsed
		// Easiest way to do this is to print the bytes first followed by a space followed by the interface name
		device_name := direntry.Name()
		if strings.Contains(device_name, "\n") {
			panic("Interface name contains newline character.")
		}
		tx_bytes := read_number_from_file_panic_on_err("/sys/class/net/" + device_name + "/statistics/tx_bytes")
		rx_bytes := read_number_from_file_panic_on_err("/sys/class/net/" + device_name + "/statistics/rx_bytes")
		// Now print the tx_bytes followed by device_name
		fmt.Fprintf(&sb, "tx_bytes: %d device_name: %s\n", tx_bytes, device_name)
		fmt.Fprintf(&sb, "rx_bytes: %d device_name: %s\n", rx_bytes, device_name)
		total_tx_bytes += tx_bytes
		total_all_bytes += tx_bytes + rx_bytes
	}

	bm.total_tx_bytes.Store(total_tx_bytes)
	bm.total_all_bytes.Store(total_all_bytes)
	fmt.Println("bm.GetTotalAllBytes:", bm.GetTotalAllBytes())
	fmt.Println("bm.GetTotalTXBytes:", bm.GetTotalTXBytes())
}
