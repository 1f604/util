// We use byte instead of rune because this alphabet contains only printable ASCII characters.
package util

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"reflect"
	"slices"
	"strings"
)

const BASE53_ALPHABET_SIZE = 53

/* Custom error types */
type Base53ErrorStrWithoutCsumTooShort struct{}

func (e Base53ErrorStrWithoutCsumTooShort) Error() string {
	return "Base53: Input string without checksum is too short (1 char min)"
}

type Base53ErrorStrWithoutCsumTooLong struct{}

func (e Base53ErrorStrWithoutCsumTooLong) Error() string {
	return "Base53: Input string without checksum is too long (51 chars max)"
}

type Base53ErrorIllegalCharacter struct{}

func (e Base53ErrorIllegalCharacter) Error() string {
	return "Base53: Input string contains illegal character"
}

type Base53ErrorIllegalPair struct{}

func (e Base53ErrorIllegalPair) Error() string {
	return "Base53: Input string contains illegal pair"
}

type Base53ErrorChecksumMismatch struct{}

func (e Base53ErrorChecksumMismatch) Error() string {
	return "Base53: Input checksum does not match string"
}

type Base53IDManager struct {
	legal_alphabet           *[]byte
	legal_alphabet_without_v *[]byte
	legal_alphabet_without_V *[]byte
	legal_alphabet_without_n *[]byte
	remapping_table          map[byte]byte
	char_to_num              []int
	num_to_char              []byte
	next_char                []byte
	illegal_pairs            *[]string
	p                        int
}

func add_range(start byte, end byte, in_arr *[]byte) {
	for i := start; i <= end; i++ {
		*in_arr = append(*in_arr, i)
	}
}

func create_alpha_without_letter(alphabet *[]byte, letter byte) []byte {
	new_alphabet := make([]byte, 0)
	for _, c := range *alphabet {
		if c != letter {
			new_alphabet = append(new_alphabet, c)
		}
	}
	return new_alphabet
}

// pregenerate means strings up to n characters will be pre-generated and stored in RandomBags for fast PopRandom and Push later.
func NewBase53IDManager() *Base53IDManager {
	all_letters_and_digits := []byte{}
	// construct letters and digits
	add_range('0', '9', &all_letters_and_digits)
	add_range('A', 'Z', &all_letters_and_digits)
	add_range('a', 'z', &all_letters_and_digits)
	// remove illegal characters
	illegal_chars := []byte{'O', '9', '1', 'I', 'l', 'W', 'w', 'm', 'd'}
	illegal_pairs := []string{"VV", "vv", "rn", "nn"}

	legal_alphabet := []byte{}
	for _, c := range all_letters_and_digits {
		if !slices.Contains(illegal_chars, c) {
			legal_alphabet = append(legal_alphabet, c)
		}
	}
	// check alphabet
	expected_alphabet := []byte{'0', '2', '3', '4', '5', '6', '7', '8', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'L', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'X', 'Y', 'Z', 'a', 'b', 'c', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'x', 'y', 'z'}
	if !reflect.DeepEqual(legal_alphabet, expected_alphabet) {
		panic("alphabet not as expected")
	}

	// check size of alphabet
	if len(legal_alphabet) != BASE53_ALPHABET_SIZE {
		panic("Expected length of alphabet to be 53")
	}

	// create alphabet without v
	legal_alphabet_without_v := create_alpha_without_letter(&legal_alphabet, 'v')

	// create alphabet without V
	legal_alphabet_without_V := create_alpha_without_letter(&legal_alphabet, 'V')

	// create alphabet without n
	legal_alphabet_without_n := create_alpha_without_letter(&legal_alphabet, 'n')

	// create remapping table
	remapping_table := map[byte]byte{
		'O': '0',
		'9': 'g',
	}
	// legal_alphabet is already ordered by construction
	// Now create num_to_char and char_to_num
	char_to_num := make([]int, 300)
	num_to_char := make([]byte, 300)
	next_char := make([]byte, 300)
	for i, c := range legal_alphabet {
		char_to_num[c] = i
		num_to_char[i] = c
		if i < len(legal_alphabet)-1 {
			next_char[c] = legal_alphabet[i+1]
		} else {
			if c != 'z' {
				panic("Unexpected char")
			}
			next_char[c] = '0'
		}
	}

	/*
			char_to_num = {ordered_alphabet[i]:i for i in range(len(alphabet))}
		num_to_char = {v:k for k,v in char_to_num.items()}

	*/
	//fmt.Println(string(all_letters_and_digits))
	//fmt.Println(string(legal_alphabet))
	//fmt.Println(string(legal_alphabet_without_v))
	//fmt.Println(string(legal_alphabet_without_V))
	//fmt.Println(string(legal_alphabet_without_n))
	//fmt.Println(remapping_table)
	//fmt.Println(char_to_num)
	//fmt.Println(num_to_char)

	return &Base53IDManager{
		legal_alphabet:           &legal_alphabet,
		legal_alphabet_without_v: &legal_alphabet_without_v,
		legal_alphabet_without_V: &legal_alphabet_without_V,
		legal_alphabet_without_n: &legal_alphabet_without_n,
		remapping_table:          remapping_table,
		char_to_num:              char_to_num,
		num_to_char:              num_to_char,
		next_char:                next_char,
		illegal_pairs:            &illegal_pairs,
		p:                        BASE53_ALPHABET_SIZE,
	}
}

type ValidationResult struct {
	Success bool
	Message string
}

// inspired by StripeIntentParams
type NewBase53IDParams struct {
	Str_without_csum string
	Csum             byte
	Remap            bool
}

func (b53m *Base53IDManager) _b53_remap(input_str string) string {
	var sb strings.Builder
	for _, c := range []byte(input_str) {
		remapped_char, ok := b53m.remapping_table[c]
		if ok {
			sb.WriteByte(remapped_char)
		} else {
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

func (b53m *Base53IDManager) _b53_check_for_illegal_chars_and_pairs(input_str string) error {
	// check for illegal characters
	for _, b := range []byte(input_str) {
		if !slices.Contains(*b53m.legal_alphabet, b) {
			return Base53ErrorIllegalCharacter{}
		}
	}
	// check for illegal pairs
	for _, illegal_pair := range *b53m.illegal_pairs {
		if strings.Contains(input_str, illegal_pair) {
			return Base53ErrorIllegalPair{}
		}
	}
	return nil
}

func (b53m *Base53IDManager) _b53_calculate_checksum(input_str string) byte {
	total := 0
	for i, c := range []byte(input_str) {
		multiplier := b53m.p - i - 2
		num := b53m.char_to_num[c]
		total += multiplier * num
	}
	checksum := total % b53m.p
	return b53m.num_to_char[checksum]
}

// See https://stackoverflow.com/questions/57993809/how-to-hide-the-default-type-constructor-in-golang
type Base53ID interface {
	GetStrWithoutCsum() string
	GetCsum() byte
	GetCombinedString() string
	Length() int
}

type _base53ID_impl struct {
	Str_without_csum string
	Csum             byte
}

func (impl *_base53ID_impl) GetStrWithoutCsum() string {
	return impl.Str_without_csum
}
func (impl *_base53ID_impl) GetCsum() byte {
	return impl.Csum
}

func (impl *_base53ID_impl) GetCombinedString() string {
	return impl.Str_without_csum + string(impl.Csum)
}

func (impl *_base53ID_impl) Length() int {
	return len(impl.Str_without_csum) + 1
}

func (impl *_base53ID_impl) String() string {
	return fmt.Sprintf("Base53ID: %s (str:%s, csum:%s)", impl.GetCombinedString(), impl.Str_without_csum, string(impl.Csum))
}

// Construction is validation.
func (b53m *Base53IDManager) NewBase53ID(str_without_csum string, csum byte, remap bool) (*_base53ID_impl, error) {
	// 1. Check lengths
	if len(str_without_csum) == 0 {
		return nil, Base53ErrorStrWithoutCsumTooShort{}
	}
	if len(str_without_csum) > 50 {
		return nil, Base53ErrorStrWithoutCsumTooLong{}
	}
	// 2. Remap if remapping is specified
	if remap { // remap both the string as well as the checksum
		str_without_csum = b53m._b53_remap(str_without_csum)
		csum = b53m._b53_remap(string(csum))[0]
	}
	// 3. Check for illegal characters
	err := b53m._b53_check_for_illegal_chars_and_pairs(str_without_csum + string(csum))
	if err != nil {
		return nil, err
	}
	// 4. Check the checksum
	recalculated_csum := b53m._b53_calculate_checksum(str_without_csum)
	if recalculated_csum != csum {
		return nil, Base53ErrorChecksumMismatch{}
	}

	return &_base53ID_impl{
		Str_without_csum: str_without_csum,
		Csum:             csum,
	}, nil
}

func (b53m *Base53IDManager) _b53_generate_random_unchecksummed(n int) (string, error) {
	var prev_char byte
	var sb strings.Builder
	for i := 0; i < n; i++ {
		var choices *[]byte
		switch prev_char {
		case 'v':
			choices = b53m.legal_alphabet_without_v
		case 'V':
			choices = b53m.legal_alphabet_without_V
		case 'n', 'r':
			choices = b53m.legal_alphabet_without_n
		default:
			choices = b53m.legal_alphabet
		}
		prev_char, err := Crypto_Random_Choice(choices)
		if err != nil {
			return "", err
		}
		sb.WriteByte(prev_char)
	}
	return sb.String(), nil
}

func (b53m *Base53IDManager) B53_generate_random_Base53ID(n int) (Base53ID, error) {
	if n < 2 {
		return nil, errors.New("Requested string too short!")
	}
	for i := 0; i < 100; i++ { // try 100 times to generate new ID, then give up
		str_without_csum, err := b53m._b53_generate_random_unchecksummed(n - 1)
		if err != nil {
			return nil, err
		}
		csum := b53m._b53_calculate_checksum(str_without_csum)
		combined := str_without_csum + string(csum)
		err = b53m._b53_check_for_illegal_chars_and_pairs(combined)
		if err == nil {
			return b53m.NewBase53ID(str_without_csum, csum, false)
		}
		// Probability of failing more than twice is extremely low.
		// Probability of failing decreases exponentially with number of attempts.
		// Failing 100 times in a row should be impossible.
	}
	log.Fatal("Fatal error: Failed 100 times in a row to generate a valid random ID.")
	panic("This should never happen.")
}

func (b53m *Base53IDManager) _increment_base53_string(str_without_csum string) string {
	n := len(str_without_csum)
	s := str_without_csum
	// compute the numerical sum
	total := 0
	reversed := ReverseString(s)
	for i, c := range []byte(reversed) {
		total += b53m.char_to_num[c] * Power_Naive(b53m.p, i)
	}
	total++
	// convert sum back into character
	var sb strings.Builder
	power := 0
	var remainder int
	for total > 0 {
		power++
		total, remainder = Divmod(total, b53m.p)
		sb.WriteByte(b53m.num_to_char[remainder])
	}
	// prepend 0s as necessary
	diff := n - sb.Len()
	if diff > 0 {
		for i := 0; i < diff; i++ {
			sb.WriteByte('0')
		}
	}
	str := sb.String()
	return ReverseString(str)
}

func (b53m *Base53IDManager) B53_generate_next_Base53ID(old_id Base53ID) (Base53ID, error) {
	// length of output shall be equal to or greater than length of input
	// 00 -> 02, 002 -> 003, 0004 -> 0005
	// special case rollover: z -> 00, zz -> 000, zzz -> 0000
	// calculate the numerical equivalent of the ID
	str_without_csum := old_id.GetStrWithoutCsum()
	//fmt.Println(str_without_csum)
	// if all z's, then roll over to the next 0s
	all_zs := true
	for _, c := range str_without_csum {
		if c != 'z' {
			all_zs = false
		}
	}

	if all_zs {
		new_string := strings.Repeat("0", len(str_without_csum)+1)
		csum := b53m._b53_calculate_checksum(new_string)
		return b53m.NewBase53ID(new_string, csum, false)
	}

	var err error
	// try 5 times before giving up
	new_string := str_without_csum
	for i := 0; i < 5; i++ {
		new_string = b53m._increment_base53_string(new_string)
		// Go through string checking for illegal pairs
		// If there is an illegal pair in new_string, increment it.
		for i := 0; i < len(new_string)-1; i++ {
			// it should look something like this:
			// hagsf7465vv00000000000
			// the disallowed pairs are: ["VV", "vv", "rn", "nn"]
			// which does not include 'zz'
			// this is good because only 'zz' rolls over to '000'
			// so incrementing an illegal pair will still result in a pair, not a triple
			// so we can edit just the illegal pair in the string without changing the rest of the string
			if slices.Contains(*b53m.illegal_pairs, new_string[i:i+2]) {
				new_pair := b53m._increment_base53_string(new_string[i : i+2])
				// mutate the string
				mutable := []byte(new_string)
				mutable[i] = new_pair[0]
				mutable[i+1] = new_pair[1]
				new_string = string(mutable)
				break
			}
		}
		// now validate it. It should be okay.
		csum := b53m._b53_calculate_checksum(new_string)
		// now check if last 2 characters form an illegal pair
		last_2_chars := []byte{new_string[len(new_string)-1], csum}
		if slices.Contains(*b53m.illegal_pairs, string(last_2_chars)) {
			// try again
			continue
		}

		var new_id Base53ID
		new_id, err = b53m.NewBase53ID(new_string, csum, false)
		if err == nil {
			return new_id, nil
		}
		fmt.Print("This should NEVER happen:", err)
		log.Fatal("This should NEVER happen:", err)
		panic(err)
	}
	fmt.Println("Failed to generate next BASE 53 ID. This should never happen")
	log.Println("Failed to generate next BASE 53 ID. This should never happen")
	return nil, err
}

// Generate all IDs of length n
func (b53m *Base53IDManager) B53_generate_all_Base53IDs(n int) ([]Base53ID, error) {
	if n < 2 {
		return nil, errors.New("Error: minimum id length is 2")
	}
	var result []Base53ID
	var cur_id Base53ID
	var err error

	cur_id, err = b53m.NewBase53ID(strings.Repeat("0", n-1), '0', false)
	if err != nil {
		return nil, err
	}

	for cur_id.Length() == n {
		result = append(result, cur_id)
		cur_id, err = b53m.B53_generate_next_Base53ID(cur_id)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (b53m *Base53IDManager) B53_generate_all_Base53IDs_int64(n int) ([]uint64, error) {
	if n < 2 {
		return nil, errors.New("Error: minimum id length is 2")
	}
	result := make([]uint64, Power_Naive(b53m.p, n-1))
	var cur_id Base53ID
	var err error

	cur_id, err = b53m.NewBase53ID(strings.Repeat("0", n-1), '0', false)
	if err != nil {
		return nil, err
	}

	for i := 0; cur_id.Length() == n; i++ {
		b := make([]byte, 8)
		copy(b, []byte(cur_id.GetCombinedString()))

		x1 := binary.LittleEndian.Uint64(b)

		result[i] = x1
		cur_id, err = b53m.B53_generate_next_Base53ID(cur_id)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (b53m *Base53IDManager) B53_generate_all_Base53IDs_int64_test(n int) ([]uint64, error) {
	result := make([]uint64, Power_Naive(b53m.p, n-1))
	for i := 0; i < len(result); i++ {
		result[i] = uint64(i)
	}

	return result, nil
}

type ShouldBase53IDBePlacedIntoSliceFn func(string) bool

// Doesn't push it into slice if it's already in map.
func (b53m *Base53IDManager) B53_generate_all_Base53IDs_int64_optimized(n int, should_be_added_fn ShouldBase53IDBePlacedIntoSliceFn) ([]uint64, error) {
	// preliminary checks
	if n < 2 {
		return nil, errors.New("Error: minimum id length is 2")
	}

	// pre-allocate array for performance. This is more storage than we need.
	result := make([]uint64, Power_Naive(b53m.p, (n-1)))

	str_without_csum_length := n - 1
	buf := make([]byte, 8) // we need 8 bytes for converting into int64, even though we can only use 7 actually since the 8th digit is reserved for checksum.
	// initialize the array
	for i := 0; i < str_without_csum_length; i++ {
		buf[i] = '0'
	}
	var cur_csum byte

	kk := -1
	for {
		str_without_csum := string(buf[0:str_without_csum_length])
		//fmt.Println(str_without_csum_length)
		// put csum into buf
		cur_csum = b53m._b53_calculate_checksum(str_without_csum) // could inline this maybe might make things faster? Could try.
		//fmt.Println("cur_csum", cur_csum)
		//fmt.Println(kk, "buf:", buf)
		buf[str_without_csum_length] = cur_csum
		// don't add buf to the result array if it's already in the map
		if should_be_added_fn != nil { // if a map is provided, then change the behavior depending on if the map contains the item or not
			item_str := string(buf[0 : str_without_csum_length+1])
			should_be_added := should_be_added_fn(item_str)
			if should_be_added { // if buf is already in the map then don't add it to the slice
				goto add_buf_to_slice
			} else { // if buf is not in the map then don't change the default behavior
				goto dont_add_buf_to_slice
			}
		}

	add_buf_to_slice:
		// We use Big Endian because it's more intuitive. I don't think this has much performance impact.
		kk++
		result[kk] = binary.BigEndian.Uint64(buf)

	dont_add_buf_to_slice: // skip the part where we add buf to the results array, instead just increment the buf again
		// generate next ID and check its length

		// length of output shall be equal to or greater than length of input
		// 00 -> 02, 002 -> 003, 0004 -> 0005
		// special case rollover: z -> 00, zz -> 000, zzz -> 0000
		// calculate the numerical equivalent of the ID
		// fmt.Println(str_without_csum)
		// if all z's, then roll over to the next 0s
		all_zs := true
		for _, c := range str_without_csum {
			if c != 'z' {
				all_zs = false
				break // early break for slight performance gain
			}
		}

		if all_zs { // we reached the end, so return the result
			//fmt.Print("Reached the end!")
			// truncate the slice to only existing elements
			return result[0 : kk+1], nil
		}

		// try to increment string 5 times before giving up
		for xx := 0; xx < 10; xx++ {
			if xx == 5 {
				panic("Tried to increment string 5 times!")
			}
			// a more efficient implementation of _increment_base53_string(str_without_csum)
			var j int // it's very important for the loop variable to be signed
			for j = str_without_csum_length - 1; j > -1; j-- {
				// increment buf[j] to next character in alphabet
				buf[j] = b53m.next_char[buf[j]]
				if buf[j] != '0' { // no overflow, so stop.
					break
				}
			}
			//fmt.Println("b53m.next_char:", b53m.next_char, b53m.next_char[buf[j]])
			//fmt.Println("Buf is now:", buf)

			// Go through string checking for illegal pairs
			// If there is an illegal pair in new_string, increment it.
			for t := 0; t < str_without_csum_length-1; t++ {
				// it should look something like this:
				// hagsf7465vv00000000000
				// the disallowed pairs are: ["VV", "vv", "rn", "nn"]
				// which does not include 'zz'
				// this is good because only 'zz' rolls over to '000'
				// so incrementing an illegal pair will still result in a pair, not a triple
				// so we can edit just the illegal pair in the string without changing the rest of the string
				if slices.Contains(*b53m.illegal_pairs, string(buf[t:t+2])) {
					//str_to_inc := string(buf[t : t+2])
					//fmt.Println("str_to_inc:", str_to_inc)
					new_pair := b53m._increment_base53_string(string(buf[t : t+2]))
					//fmt.Println("new pair:", new_pair)
					// mutate the string
					buf[t] = new_pair[0]
					buf[t+1] = new_pair[1]
					//fmt.Println("MUTATED NEW BUF IS:", buf)
					break
				}
			}
			csum := b53m._b53_calculate_checksum(string(buf[0:str_without_csum_length]))
			// now check if last 2 characters form an illegal pair
			last_2_chars := []byte{buf[str_without_csum_length-1], csum}
			if slices.Contains(*b53m.illegal_pairs, string(last_2_chars)) {
				// try again
				continue
			}
			break
		}
	}

	return result, nil
}

func (b53m *Base53IDManager) Convert_uint64_to_Base53ID(bigendian_uint64 uint64, length int) (*_base53ID_impl, error) {
	buf := make([]byte, 8) //nolint:gomnd // 8 is size of int64
	binary.BigEndian.PutUint64(buf, bigendian_uint64)
	return b53m.NewBase53ID(string(buf[0:length-1]), buf[length-1], false)
}

func (b53m *Base53IDManager) Convert_uint64_to_byte_array(bigendian_uint64 uint64) []byte {
	buf := make([]byte, 8) //nolint:gomnd // 8 is size of int64
	binary.BigEndian.PutUint64(buf, bigendian_uint64)
	return buf
}

func Convert_uint64_to_str(bigendian_uint64 uint64, length int) string {
	buf := make([]byte, 8) //nolint:gomnd // 8 is size of int64
	binary.BigEndian.PutUint64(buf, bigendian_uint64)
	return string(buf[0:length])
}

func Convert_str_to_uint64(input_str string) uint64 {
	buf := make([]byte, 8) //nolint:gomnd // 8 is size of int64
	copy(buf, []byte(input_str))
	return binary.BigEndian.Uint64(buf)
}
