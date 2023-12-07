package util

import (
	"fmt"
	"math/rand"
	"strings"
)

var illegalCharacters = []rune{'O', '9', '1', 'I', 'l', 'W', 'w', 'm', 'd'}
var illegalPairs = []string{"VV", "vv", "rn", "nn"}
var alphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var p = len(alphabet)

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i < n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func checkPrime() {
	if !isPrime(p) {
		panic("size of alphabet is not prime")
	}
	if p != 53 {
		panic("size of alphabet is not 53")
	}
	if p != len(removeDuplicates(alphabet)) {
		panic("alphabet contains duplicates")
	}
}

func removeDuplicates(arr []rune) []rune {
	encountered := map[rune]bool{}
	result := []rune{}
	for _, val := range arr {
		if !encountered[val] {
			encountered[val] = true
			result = append(result, val)
		}
	}
	return result
}

func remap(s string) string {
	ls := []rune(s)
	for i := 0; i < len(ls); i++ {
		if ls[i] == 'O' {
			ls[i] = '0'
		} else if ls[i] == '9' {
			ls[i] = 'g'
		}
	}
	return string(ls)
}

func generateRandomUnchecksummed(n int) string {
	prevChar := ' '
	result := []rune{}
	for i := 0; i < n; i++ {
		if prevChar == 'v' {
			choices := removeCharacter(alphabet, 'v')
			prevChar = choices[rand.Intn(len(choices))]
		} else if prevChar == 'V' {
			choices := removeCharacter(alphabet, 'V')
			prevChar = choices[rand.Intn(len(choices))]
		} else if prevChar == 'n' || prevChar == 'r' {
			choices := removeCharacter(alphabet, 'n')
			prevChar = choices[rand.Intn(len(choices))]
		} else {
			prevChar = alphabet[rand.Intn(len(alphabet))]
		}
		result = append(result, prevChar)
	}
	return string(result)
}

func removeCharacter(arr []rune, char rune) []rune {
	result := []rune{}
	for _, val := range arr {
		if val != char {
			result = append(result, val)
		}
	}
	return result
}

func checkForIllegalCharsAndPairs(s string) bool {
	for _, c := range s {
		if !contains(alphabet, c) {
			return false
		}
	}
	for _, ip := range illegalPairs {
		if strings.Contains(s, ip) {
			return false
		}
	}
	return true
}

func contains(arr []rune, char rune) bool {
	for _, val := range arr {
		if val == char {
			return true
		}
	}
	return false
}

func getChecksum(s string) rune {
	total := 0
	for i, c := range s {
		multiplier := p - 1 - i
		num := getCharToNum(c)
		total += multiplier * num
	}
	checksum := total % p
	return getNumToChar(checksum)
}

func getCharToNum(c rune) int {
	for i, val := range alphabet {
		if val == c {
			return i
		}
	}
	return -1
}

func getNumToChar(num int) rune {
	return alphabet[num]
}

type ValidationResult struct {
	Success bool
	Message string
}

type Base53ID struct {
	StringWithoutChecksum string
	ChecksumChar          rune
}

func NewBase53ID(stringWithoutChecksum string, checksumChar rune, remap bool, autogenerate bool) (*Base53ID, error) {
	if !checkForIllegalCharsAndPairs(stringWithoutChecksum + string(checksumChar)) {
		return nil, fmt.Errorf("Invalid characters or pairs")
	}
	if autogenerate {
		checksumChar = getChecksum(stringWithoutChecksum)
	}
	if len(string(checksumChar)) != 1 {
		return nil, fmt.Errorf("Checksum must be exactly one character")
	}
	if len(stringWithoutChecksum) < 1 {
		return nil, fmt.Errorf("String too short")
	}
	if len(stringWithoutChecksum) > p-2 {
		return nil, fmt.Errorf("String too long")
	}
	if remap {
		stringWithoutChecksum = remap(stringWithoutChecksum)
		checksumChar = rune(remap(string(string(checksumChar)))[0])
	}
	recalculatedChecksum := getChecksum(stringWithoutChecksum)
	if checksumChar != recalculatedChecksum {
		return nil, fmt.Errorf("Checksum does not match")
	}
	return &Base53ID{
		StringWithoutChecksum: stringWithoutChecksum,
		ChecksumChar:          checksumChar,
	}, nil
}

func (b *Base53ID) String() string {
	return b.StringWithoutChecksum + string(b.ChecksumChar)
}

func (b *Base53ID) Equals(other *Base53ID) bool {
	if other == nil {
		return false
	}
	return b.String() == other.String()
}

func (b *Base53ID) HashCode() int {
	return hash(b.String())
}

func hash(s string) int {
	h := 0
	for _, c := range s {
		h = 31*h + int(c)
	}
	return h
}

func GenerateRandomBase53ID(n int) (*Base53ID, error) {
	for i := 0; i < 50; i++ {
		stringWithoutChecksum := generateRandomUnchecksummed(n)
		checksumChar := getChecksum(stringWithoutChecksum)
		if !containsIllegalPairs(stringWithoutChecksum + string(checksumChar)) {
			return NewBase53ID(stringWithoutChecksum, checksumChar, false, false)
		}
	}
	return nil, fmt.Errorf("Unable to generate new ID")
}

func containsIllegalPairs(s string) bool {
	for _, ip := range illegalPairs {
		if strings.Contains(s, ip) {
			return true
		}
	}
	return false
}

func GenerateNextBase53ID(oldID *Base53ID) (*Base53ID, error) {
	if oldID == nil {
		return nil, fmt.Errorf("Invalid old ID")
	}
	newS := oldID.StringWithoutChecksum
	if allZ(newS) {
		newS = strings.Repeat("0", len(newS)+1)
		checksumChar := getChecksum(newS)
		return NewBase53ID(newS, checksumChar, false, false)
	}
	for i := 0; i < 5; i++ {
		newS = incrementBase53String(newS)
		newID, err := NewBase53ID(newS, 'z', true, true)
		if err == nil {
			return newID, nil
		}
		if !strings.HasPrefix(err.Error(), "Invalid characters or pairs") {
			return nil, err
		}
		newArr := []rune(newS)
		for i := 0; i < len(newS)-1; i++ {
			pair := string(newS[i]) + string(newS[i+1])
			if contains(illegalPairs, pair) {
				nextPair := incrementBase53String(pair)
				newArr[i] = rune(nextPair[0])
				newArr[i+1] = rune(nextPair[1])
			}
		}
		newS = string(newArr)
		checksumChar := getChecksum(newS)
		newID, err = NewBase53ID(newS, checksumChar, false, false)
		if err == nil {
			return newID, nil
		}
	}
	return nil, fmt.Errorf("Unable to generate new ID")
}

func incrementBase53String(strWithoutChecksum string) string {
	n := len(strWithoutChecksum)
	s := []rune(strWithoutChecksum)
	total := 0
	for i := len(s) - 1; i >= 0; i-- {
		total += getCharToNum(s[i]) * pow(p, len(s)-1-i)
	}
	total++
	newStr := []rune{}
	power := 0
	for total > 0 {
		power++
		total, remainder := div(total, p)
		newStr = append(newStr, getNumToChar(remainder))
	}
	diff := n - len(newStr)
	if diff > 0 {
		newStr = append(newStr, []rune(strings.Repeat("0", diff))...)
	}
	return reverse(string(newStr))
}

func pow(a, b int) int {
	result := 1
	for b > 0 {
		if b&1 == 1 {
			result *= a
		}
		a *= a
		b >>= 1
	}
	return result
}

func div(a, b int) (int, int) {
	return a / b, a % b
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func allZ(s string) bool {
	for _, c := range s {
		if c != 'z' {
			return false
		}
	}
	return true
}

func main() {
	checkPrime()
	fmt.Println("All checks passed")
}
