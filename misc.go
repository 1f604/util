// Print stuff like Go's internal representation:
//	fmt.Printf("dir_names_list: %#v\n", dir_names_list)

package util

import (
	b64 "encoding/base64"
	"reflect"
	"strings"

	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"runtime"
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
	// remove leading and trailing whitespace
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}

// this function assumes file pointer is valid.
// We could probably make this more efficient by calculating the file size in-process instead of making syscall each time.
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

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func Copy_Slice_Into_150_Arr(slice []byte, arr [150]byte) {
	minlen := min(len(arr), len(slice))
	for i := 0; i < minlen; i++ {
		arr[i] = slice[i]
	}
}

// Returns random string consisting of letters and numbers
func Crypto_Rand_Base64String(length int) string {
	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		log.Fatal("error reading crypto random:", err)
		panic("error reading crypto random:" + err.Error())
	}
	// The slice should now contain random bytes instead of only zeroes.
	return b64.StdEncoding.EncodeToString(buf)
}

// Character class lookup table
var IsAlnum = [256]byte{
	/* 00-07  NUL, SOH, STX, ETX, EOT, ENQ, ACK, BEL */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* 08-0F  BS,  HT,  LF,  VT,  FF,  CR,  SO,  SI  */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* 10-17  DLE, DC1, DC2, DC3, DC4, NAK, SYN, ETB */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* 18-1F  CAN, EM,  SUB, ESC, FS,  GS,  RS,  US  */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* 21-27  SP ! " # $ % & '   */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* 28-2F   ( ) * + , - . /   */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* 30-37   0 1 2 3 4 5 6 7   */ 1, 1, 1, 1, 1, 1, 1, 1,
	/* 38-3F   8 9 : ; < = > ?   */ 1, 1, 0, 0, 0, 0, 0, 0,
	/* 40-47   @ A B C D E F G   */ 0, 1, 1, 1, 1, 1, 1, 1,
	/* 48-4F   H I J K L M N O   */ 1, 1, 1, 1, 1, 1, 1, 1,
	/* 50-57   P Q R S T U V W   */ 1, 1, 1, 1, 1, 1, 1, 1,
	/* 58-5F   X Y Z [ \ ] ^ _   */ 1, 1, 1, 0, 0, 0, 0, 0,
	/* 60-67   ` a b c d e f g   */ 0, 1, 1, 1, 1, 1, 1, 1,
	/* 68-6F   h i j k l m n o   */ 1, 1, 1, 1, 1, 1, 1, 1,
	/* 70-77   p q r s t u v w   */ 1, 1, 1, 1, 1, 1, 1, 1,
	/* 78-7F   x y z { | } ~ DEL */ 1, 1, 1, 0, 0, 0, 0, 0,
	/* 80-87 */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* 88-8B */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* 90-97 */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* 98-9F */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* A0-A7 */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* A8-AF */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* B0-B7 */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* B8-BF */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* C0-C7 */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* C8-CF */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* D0-D7 */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* D8-DF */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* E0-E7 */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* E8-EF */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* F0-F7 */ 0, 0, 0, 0, 0, 0, 0, 0,
	/* F8-FF */ 0, 0, 0, 0, 0, 0, 0, 0,
}

// Returns random string consisting of letters and numbers
func Crypto_Rand_Alnum_String(length int) string {
	// Just keep grabbing more randomness until we have enough
	// This rejection sampling approach is extremely wasteful but it's simple
	if length < 1 {
		return ""
	}
	result := make([]byte, 0, length)
	for count := 0; count < 10; count++ {
		// fmt.Println("Iterating...")
		buf := make([]byte, length)
		_, err := rand.Read(buf)
		if err != nil {
			log.Fatal("error reading crypto random:", err)
			panic("error reading crypto random:" + err.Error())
		}
		// The slice should now contain random bytes instead of only zeroes.
		b64string := b64.StdEncoding.EncodeToString(buf)
		// fmt.Println("len(b64string):", len(b64string))
		for i := range b64string {
			if IsAlnum[b64string[i]] == 1 {
				result = append(result, b64string[i])
				// fmt.Println("Adding character:", string(b64string[i]))
				if len(result) == length {
					return string(result)
				}
			}
		}
	}
	log.Fatal("Failed to generate random string")
	panic("This shouldn't happen")
}

// This function works, I've manually tested it.
// Returns integers from 0 up to AND NOT INCLUDING max
func Crypto_Randint(max int) (int, error) {
	val, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(val.Int64()), nil
}

/* Custom error types */
type CryptoRandomChoiceEmptySliceError struct{}

func (e CryptoRandomChoiceEmptySliceError) Error() string {
	return "Crypto_Random_Choice Error: Input slice is of length zero"
}

// This function works, I've manually tested it.
func Crypto_Random_Choice[T any](arr *[]T) (T, error) { //nolint:ireturn // why is this not okay
	// This function HAS TO BE generic because converting slice of interface{} is O(N) time because it has to convert every element in the slice!!!
	// return the zero value for T
	var zero_value T
	n := len(*arr)
	if n == 0 {
		return zero_value, CryptoRandomChoiceEmptySliceError{}
	}
	idx, err := Crypto_Randint(n)
	if err != nil {
		return zero_value, err
	}
	return (*arr)[idx], nil
}

// calculates a to the power of b mod m. If m is 0 then just returns a to the power of b.
// This function seems to create a memory leak, but it doesn't.
// Anyway, it's better to use custom power
func Power_Slow(a, b, m int) int {
	result := new(big.Int).Exp(
		big.NewInt(int64(a)),
		big.NewInt(int64(b)),
		big.NewInt(int64(m)),
	)
	return int(result.Int64())
}

// Naive algorithm, only suitable for small b.
func Power_Naive(a, b int) int {
	// VERY IMPORTANT special case this fucked me up good
	if b == 0 {
		return 1
	}
	multiplier := a
	for i := 1; i < b; i++ {
		a *= multiplier
	}
	return a
}

func ReverseString(s string) string {
	chars := []rune(s)
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}

func Divmod(numerator, denominator int) (int, int) {
	quotient := numerator / denominator // integer division, decimals are truncated
	remainder := numerator % denominator
	return quotient, remainder
}

func ReplaceString(str string, replacement rune, index int) string {
	out := []byte(str)
	out[index] = byte(replacement)
	return string(out)
}

type fn_type func()

// Does not tick shift - will run function precisely every X seconds even if function takes some time to run - as long as the function doesn't take too long of course.
//
// Synchronous - next call cannot start until previous call has finished.
func RunFuncEveryXSeconds(fn fn_type, run_interval_seconds int) {
	for range time.Tick(time.Second * time.Duration(run_interval_seconds)) {
		//log.Println("Running functioN!")
		fn()
	}
}

// Returns error if unix timestamp is before 2023 or after the year 20,000
//
// Otherwise returns nil
func Validate_Timestamp_Common(timestamp_unix int64) error {
	year_2023 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	year_20000 := time.Date(20000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	switch { // timestamp is either expiry time or generated time
	case timestamp_unix < year_2023:
		return fmt.Errorf("Timestamp %#v is before the year 2023", timestamp_unix)
	case timestamp_unix > year_20000:
		return fmt.Errorf("Timestamp %#v is after the year 20,000", timestamp_unix)
	}
	return nil
}

// This function works, I've manually tested it.
func IsSameType(a, b interface{}) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(b)
}
