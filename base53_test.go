package util_test

import (
	"strings"
	"testing"

	"github.com/1f604/util"
)

func expect(t *testing.T, b53m *util.Base53IDManager, input_str string, input_csum byte, remap bool, expected_result interface{}) {
	t.Helper()

	id, err := b53m.NewBase53ID(input_str, input_csum, remap)
	if err != nil {
		util.Assert_error_equals(t, err, expected_result.(string), 2)
	} else {
		util.Assert_result_equals_interface(t, id.String(), err, expected_result.(string), 2)
	}
}

func remap_helper(t *testing.T, b53m *util.Base53IDManager, input_str string, input_rune byte, output_str string) {
	t.Helper()

	result, err := b53m.NewBase53ID(input_str, input_rune, true)
	if err != nil {
		panic(err)
	}
	util.Assert_result_equals_interface(t, result.GetCombinedString(), err, output_str, 2)
}

func Test_Remapping(t *testing.T) {
	t.Parallel()

	b53m := util.NewBase53IDManager()

	// remapping flag = true, test that it remaps
	remap_helper(t, b53m, "8uO", '2', "8u02")
	remap_helper(t, b53m, "OfJy", 'O', "0fJy0") // this test checks that the remapping works for the checksum character too

	remap_helper(t, b53m, "O9vi", '9', "0gvig") // as does this one, which checks remapping for both O and 9
	remap_helper(t, b53m, "OfH9", 'U', "0fHgU")
	remap_helper(t, b53m, "O9tY", '9', "0gtYg")
	remap_helper(t, b53m, "O9r9", 'B', "0grgB")

	// remapping flag = false, test it errors
	_, err := b53m.NewBase53ID("8uO", '2', false)
	util.Assert_error_equals(t, err, "Base53: Input string contains illegal character", 1)

	_, err = b53m.NewBase53ID("8u", 'O', false)
	util.Assert_error_equals(t, err, "Base53: Input string contains illegal character", 1)

	_, err = b53m.NewBase53ID("898", '2', false)
	util.Assert_error_equals(t, err, "Base53: Input string contains illegal character", 1)
}

func Test_Base53_Validation_Fails_Illegal_Chars(t *testing.T) {
	t.Parallel()

	b53m := util.NewBase53IDManager()
	illegal_chars := []byte{'1', '9', 'I', 'O', 'l', 'W', 'w', 'd', 'm'}

	for _, ic := range illegal_chars {
		// one character test
		expect(t, b53m, string(ic), '4', false, "Base53: Input string contains illegal character")
		expect(t, b53m, "4", ic, false, "Base53: Input string contains illegal character")
		expect(t, b53m, "af4876", ic, false, "Base53: Input string contains illegal character")

		// two character test
		expect(t, b53m, string(ic)+"a", '4', false, "Base53: Input string contains illegal character")
		expect(t, b53m, "a"+string(ic), '4', false, "Base53: Input string contains illegal character")

		// three character test
		expect(t, b53m, string(ic)+"ab", '4', false, "Base53: Input string contains illegal character")
		expect(t, b53m, "ab"+string(ic), '4', false, "Base53: Input string contains illegal character")
		expect(t, b53m, "a"+string(ic)+"b", '4', false, "Base53: Input string contains illegal character")
	}
}

func Test_length_check(t *testing.T) {
	t.Parallel()

	b53m := util.NewBase53IDManager()
	expect(t, b53m, "", '4', true, "Base53: Input string without checksum is too short (1 char min)")
	expect(t, b53m, "", '4', false, "Base53: Input string without checksum is too short (1 char min)")

	expect(t, b53m, strings.Repeat("a", 51), '4', true, "Base53: Input string without checksum is too long (51 chars max)")
	expect(t, b53m, strings.Repeat("a", 52), '4', true, "Base53: Input string without checksum is too long (51 chars max)")
}

func Test_fails_illegal_pairs(t *testing.T) {
	t.Parallel()

	b53m := util.NewBase53IDManager()
	illegal_pairs := []string{"vv", "nn", "VV", "rn"}
	for _, ip := range illegal_pairs {
		// 0 char test
		expect(t, b53m, string(ip[0]), ip[1], true, "Base53: Input string contains illegal pair")
		expect(t, b53m, string(ip[0]), ip[1], false, "Base53: Input string contains illegal pair")
		expect(t, b53m, ip, '4', true, "Base53: Input string contains illegal pair")
		expect(t, b53m, ip, '4', false, "Base53: Input string contains illegal pair")

		// 1 char test
		expect(t, b53m, "a"+string(ip[0]), ip[1], true, "Base53: Input string contains illegal pair")
		expect(t, b53m, ip+"a", '4', true, "Base53: Input string contains illegal pair")
		expect(t, b53m, "a"+ip, '4', true, "Base53: Input string contains illegal pair")

		// 2 char test
		expect(t, b53m, "ab"+string(ip[0]), ip[1], true, "Base53: Input string contains illegal pair")
		expect(t, b53m, ip+"ab", '4', true, "Base53: Input string contains illegal pair")
		expect(t, b53m, "ab"+ip, '4', true, "Base53: Input string contains illegal pair")
		expect(t, b53m, "a"+ip+"b", '4', true, "Base53: Input string contains illegal pair")
	}
}

func Test_checksum_validation(t *testing.T) {
	t.Parallel()
	b53m := util.NewBase53IDManager()

	// Test checksum passes
	expect(t, b53m, "0", '0', true, "Base53ID: 00 (str:0, csum:0)")
	expect(t, b53m, "0gsk", 'i', true, "Base53ID: 0gski (str:0gsk, csum:i)")
	expect(t, b53m, "Ogsk", 'i', true, "Base53ID: 0gski (str:0gsk, csum:i)")
	expect(t, b53m, "Ogsk", 'i', false, "Base53: Input string contains illegal character")

	// Test checksum fails
	expect(t, b53m, "0", '2', true, "Base53: Input checksum does not match string")
	expect(t, b53m, "0gsk", 'j', true, "Base53: Input checksum does not match string")
	expect(t, b53m, "0g5k", 'i', true, "Base53: Input checksum does not match string")
	expect(t, b53m, "0gSk", 'i', true, "Base53: Input checksum does not match string")
	expect(t, b53m, "0gsK", 'i', true, "Base53: Input checksum does not match string")
	expect(t, b53m, "0sgk", 'i', true, "Base53: Input checksum does not match string")
	expect(t, b53m, "g0sk", 'i', true, "Base53: Input checksum does not match string")
	expect(t, b53m, "0gsi", 'k', true, "Base53: Input checksum does not match string")
}

func Test_Generate_Random_generates_all_legal_chars(t *testing.T) {
	t.Parallel()
	b53m := util.NewBase53IDManager()

	legal_alphabet := []rune{'0', '2', '3', '4', '5', '6', '7', '8', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'L', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'X', 'Y', 'Z', 'a', 'b', 'c', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'x', 'y', 'z'}

	for n := 1; n < 5; n++ {
		seen := make(map[rune]bool)
		for j := 0; j < 2000; j++ {
			rndstr, err := b53m.B53_generate_random_Base53ID(n)
			if err != nil {
				panic(err)
			}
			for _, c := range rndstr.GetCombinedString() {
				seen[c] = true
			}
		}
		for _, c := range legal_alphabet {
			_, ok := seen[c]
			if !ok {
				panic("Letter not generated: " + string(c))
			}
		}
	}
}

func Test_Generate_Random_generates_all_legal_strings(t *testing.T) {
	t.Parallel()
	b53m := util.NewBase53IDManager()

	// test it generates all legal 2 character strings
	n := 2
	results := make(map[string]bool)
	if len(results) != 0 {
		panic("Not expected length")
	}
	for i := 0; i < 50000; i++ { // should have less than 0.001% chance of failing
		rs, err := b53m.B53_generate_random_Base53ID(n)
		if err != nil {
			panic(err)
		}
		results[rs.GetCombinedString()] = true
	}
	if len(results) != 53*53-4*2 { // only 'VV' 'vv' 'rn' and 'nn' are disallowed, but these pairs can occur in two places
		panic("Unexpected number of generated strings")
	}
}

func Test_Generate_Random_generates_different_strings(t *testing.T) {
	t.Parallel()
	b53m := util.NewBase53IDManager()

	results := make(map[string]bool)
	if len(results) != 0 {
		panic("Not expected length")
	}
	for i := 0; i < 1000; i++ {
		rs, err := b53m.B53_generate_random_Base53ID(8)
		if err != nil {
			panic(err)
		}
		results[rs.GetCombinedString()] = true
	}
	if len(results) != 1000 { // should never get repeats.
		panic("Unexpected number of generated strings")
	}
}

func Test_Generate_Next_generates_right_number_of_strings(t *testing.T) {
	t.Parallel()
	b53m := util.NewBase53IDManager()

	count := map[int]int{
		1: 53,
		2: 53*53 - 4*2,  // this is the number of valid strings after removing those with illegal pairs
		3: 148457 - 208, // I don't know how I calculated this number.
	}

	for i := 1; i < 4; i++ {
		var id util.Base53ID
		id, err := b53m.NewBase53ID(strings.Repeat("0", i), '0', false)
		util.Check_err(err)
		results := map[string]bool{
			id.GetCombinedString(): true,
		}
		num_strings := count[i] // subtract the number of illegal pairs
		for x := 1; x < num_strings; x++ {
			id, err = b53m.B53_generate_next_Base53ID(id)
			util.Check_err(err)
			sid := id.GetCombinedString()
			util.Assert_result_equals_interface(t, len(sid), err, i+1, 1)
			results[sid] = true
		}
		util.Assert_result_equals_interface(t, len(results), err, num_strings, 1)
	}
}

func Test_Generate_all_generates_right_number_of_strings(t *testing.T) {
	t.Parallel()
	b53m := util.NewBase53IDManager()

	count := map[int]int{
		2: 53,
		3: 53*53 - 4*2,  // this is the number of valid strings after removing those with illegal pairs
		4: 148457 - 208, // I don't know how I calculated this number.
	}

	for i := 2; i < 5; i++ {
		ids, err := b53m.B53_generate_all_Base53IDs(i)
		util.Check_err(err)

		size := len(ids)
		expected_num_strings := count[i] // subtract the number of illegal pairs
		util.Assert_result_equals_interface(t, size, err, expected_num_strings, 1)

		results := map[string]bool{}
		for _, id := range ids {
			results[id.GetCombinedString()] = true
			util.Assert_result_equals_interface(t, len(id.GetCombinedString()), err, i, 1)
		}
		util.Assert_result_equals_interface(t, len(results), err, expected_num_strings, 1)
	}
}

func Test_Generate_all_uint64_optimized_generates_strings_with_trailing_zeroes(t *testing.T) {
	t.Parallel()
	b53m := util.NewBase53IDManager()

	_, err := b53m.B53_generate_all_Base53IDs_int64_optimized(1)
	util.Assert_error_equals(t, err, "Error: minimum id length is 2", 1)

	for i := 2; i < 5; i++ {
		results, err := b53m.B53_generate_all_Base53IDs_int64_optimized(i)
		util.Assert_no_error(t, err, 1)
		for _, num := range results {
			arr := b53m.Convert_uint64_to_byte_array(num)
			for j := i; j < 8; j++ {
				util.Assert_result_equals_interface(t, arr[j], err, byte(0), 1)
			}
		}
	}
}

func Test_Generate_all_uint64_optimized_generates_exact_same_strings(t *testing.T) {
	t.Parallel()
	b53m := util.NewBase53IDManager()

	for n := 2; n < 5; n++ {
		expected_ids, err := b53m.B53_generate_all_Base53IDs(n)
		util.Check_err(err)
		actual_ids, err := b53m.B53_generate_all_Base53IDs_int64_optimized(n)
		util.Check_err(err)

		// check lengths are equal
		util.Assert_result_equals_interface(t, len(actual_ids), err, len(expected_ids), 1)

		// check each string is equal
		for i := 0; i < len(actual_ids); i++ {
			expected_str := expected_ids[i].GetCombinedString()
			actual_str := b53m.Convert_uint64_to_str(actual_ids[i], n)
			util.Assert_result_equals_interface(t, actual_str, err, expected_str, 1)
		}
	}
}
