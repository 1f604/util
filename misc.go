// Print stuff like Go's internal representation:
//	fmt.Printf("dir_names_list: %#v\n", dir_names_list)

package util

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type retrylib_task func()

type retrylib_counter struct {
	mu               sync.Mutex
	max_val          time.Duration
	private_variable time.Duration
}

func (c *retrylib_counter) incr() {
	c.mu.Lock()
	if c.private_variable < c.max_val {
		c.private_variable += time.Second
	}
	c.mu.Unlock()
}

func (c *retrylib_counter) getValue() time.Duration {
	c.mu.Lock()
	n := c.private_variable
	c.mu.Unlock()

	return n
}

func (c *retrylib_counter) zero() {
	c.mu.Lock()
	c.private_variable = 0
	c.mu.Unlock()
}

func newRetrylibCounter(maxval time.Duration) *retrylib_counter {
	return &retrylib_counter{max_val: maxval}
}

func Retryfunc(taskname string, dotask retrylib_task, expected_duration time.Duration, max_wait time.Duration) {
	count := newRetrylibCounter(max_wait)
	for {
		start := time.Now()
		dotask()
		duration := time.Since(start)
		log.Printf("%s finished after %d seconds.\n", taskname, duration/time.Second)

		if duration > expected_duration {
			count.zero()
		} else {
			count.incr()
		}
		log.Printf("%s: sleeping for %d seconds before re-running\n", taskname, count.getValue()/time.Second)
		time.Sleep(count.getValue())
	}
}

func Retryproc(procname string, expected_duration time.Duration, max_wait time.Duration) {
	f := func() {
		cmd := exec.Command(procname)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		log.Printf("launching process %s ...\n", procname)
		err := cmd.Run()

		if err != nil {
			log.Printf("process %s: an error occurred: %v\n", procname, err)
		} else {
			log.Printf("process %s completed without error.\n", procname)
		}
	}
	Retryfunc("command "+procname, f, expected_duration, max_wait)
}

// https://stackoverflow.com/questions/19965795/how-to-write-log-to-file
func SetLogFile(filename string) *os.File {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}

	log.SetOutput(f)
	log.Println("Started logging.")
	return f
}

// https://stackoverflow.com/questions/21743841/how-to-avoid-annoying-error-declared-and-not-used
func Use(vals ...interface{}) {
	for _, val := range vals {
		_ = val
	}
}

func Int64_to_string(num int64) string {
	return strconv.FormatInt(num, 10)
}

func String_to_int64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// this function assumes file pointer is valid.
func Get_file_size(f *os.File) int64 {
	fi, err := f.Stat()
	if err != nil {
		panic("ERROR: STAT ON LOG FILE FAILED!!!")
	}
	return fi.Size()
}

func Check_err(err error) {
	if err != nil {
		panic(err)
	}
}

func BuildStruct[T any]() *T {
	return new(T)
}

func Getxattr(path string, name string, data []byte) (int, error) {
	return unix.Getxattr(path, name, data)
}

func Setxattr(path string, name string, data []byte, flags int) error {
	return unix.Setxattr(path, name, data, flags)
}
