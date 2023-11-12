package util

import (
	"log"
	"net"
	"strconv"
)

func Check_no_other_instances_running(socket_addr string) {
	/*  This implementation uses Linux abstract domain sockets, which is a Linux-specific feature.

	    A cross-platform approach would be to use TCP sockets instead, but that uses an entire TCP port, which is a limited resource.

	    For a discussion of the merits of flock vs abstract sockets see this:
	    	https://blog.petrzemek.net/2017/07/24/ensuring-that-a-linux-program-is-running-at-most-once-by-using-abstract-sockets/

		For a discussion of the problems with pidfiles see this:
			https://stackoverflow.com/questions/25906020/are-pid-files-still-flawed-when-doing-it-right
	*/
	_, err := net.Listen("unix", socket_addr)
	if err != nil {
		log.Fatal("Error: another instance of this program is already running.")
	}
}

// https://rosettacode.org/wiki/Determine_if_only_one_instance_is_running#Port
func CheckTCPPort(port int) net.Listener {
	// This is the TCP version
	var l net.Listener
	var err error
	if l, err = net.Listen("tcp", ":"+strconv.Itoa(port)); err != nil {
		log.Fatal("Error: another instance of this program is already running.")
	}
	return l
}
