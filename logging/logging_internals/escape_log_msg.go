package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/1f604/util"
)

func Encode_log_msg(dangerous_bytes []byte) []byte {
	// It's probably fine to use strconv.QuoteASCII or fmt.Printf("%q")
	// But I'm extremely paranoid and I have to define my own whitelist
	var sb strings.Builder
	// Check each rune
	dangerous_str := string(dangerous_bytes)  // this cannot fail
	for _, char_rune := range dangerous_str { // this cannot fail, it will just produce the special INVALID UTF-8 character
		// check if rune is alphanumeric, if not we convert it to hex code
		// check if ASCII printable.
		// range 32 to 126 inclusive contains all ASCII printable characters from space to tilde
		/*
			From the Go Language Spec:

				Several backslash escapes allow arbitrary values to be encoded as ASCII text.
				There are four ways to represent the integer value as a numeric constant:
				\x followed by exactly two hexadecimal digits;
				\u followed by exactly four hexadecimal digits;
				\U followed by exactly eight hexadecimal digits,
				and a plain backslash \ followed by exactly three octal digits.
				In each case the value of the literal is the value represented by the digits in the corresponding base.

				Although these representations all result in an integer, they have different valid ranges.
				Octal escapes must represent a value between 0 and 255 inclusive.
				Hexadecimal escapes satisfy this condition by construction.
				The escapes \u and \U represent Unicode code points so within them
				some values are illegal, in particular those above 0x10FFFF and surrogate halves.
		*/
		switch {
		case char_rune >= 32 && char_rune <= 126 && char_rune != 92: // ASCII code 92 is backslash, which we want to escape.
			_, err := sb.WriteRune(char_rune)
			util.Check_err(err)
		case char_rune < 128: //nolint:gomnd // if it's ASCII then we can represent it with hex code
			_, err := sb.WriteString(fmt.Sprintf("\\x%02x", char_rune))
			util.Check_err(err)
		default: // otherwise use full unicode representation
			_, err := sb.WriteString(fmt.Sprintf("\\U%08x", char_rune))
			util.Check_err(err)
		}
	}
	return []byte(sb.String())
}

// Do not call this method on a running server as it may crash the server
// Even if you call it on your own log file, if your log file was partly written it can still crash your program
// But the Rotate function isn't affected because it only looks for newlines.
// Since we don't write any newlines except at the end of every log message, it's fine.
func Decode_log_msg(msg_bytes []byte) string {
	// This function is not secure, only pass into it log files that you generated yourself.
	// It will just silently fail on messages that aren't properly encoded.
	var sb strings.Builder
	for i := 0; i < len(msg_bytes); i++ {
		if msg_bytes[i] == 92 { //nolint:gomnd // we all know 92 is backslash...
			switch {
			case msg_bytes[i+1] == 'x':
				char_int, err := strconv.ParseInt(string(msg_bytes[i+2:i+4]), 16, 64)
				util.Check_err(err)
				_, err = sb.WriteRune(rune(char_int))
				util.Check_err(err)
				i += 1 + 2 //nolint:gomnd // skip pointer by 1+2 because it's encoded as \x02
			case msg_bytes[i+1] == 'U':
				char_int, err := strconv.ParseInt(string(msg_bytes[i+2:i+10]), 16, 64)
				util.Check_err(err)
				_, err = sb.WriteRune(rune(char_int))
				util.Check_err(err)
				i += 1 + 8 //nolint:gomnd // skip pointer by 1+8 because it's encoded as \U08
			default:
				panic("Decode_log_msg: Unexpected input")
			}
		} else {
			err := sb.WriteByte(msg_bytes[i])
			util.Check_err(err)
		}
	}
	return sb.String()
}
