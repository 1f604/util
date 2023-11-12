package util_test

import (
	"testing"

	"github.com/1f604/util"

	logging_internals "github.com/1f604/util/logging/logging_internals"
)

func Test_Encode_log_msg(t *testing.T) {
	t.Parallel()

	// empty string
	const empty_str = ``
	encoded := logging_internals.Encode_log_msg([]byte(empty_str))
	util.Assert_result_equals_bytes(t, encoded, nil, empty_str, 1)

	// simple hello world with some symbols
	const simple_ascii_str = `hello world! @~/"~~!!'"!`
	encoded = logging_internals.Encode_log_msg([]byte(simple_ascii_str))
	util.Assert_result_equals_bytes(t, encoded, nil, simple_ascii_str, 1)

	// example nginx log message
	const example_nginx_log_line = `66.249.65.159 - - [06/Nov/2014:19:10:38 +0600] "GET /news/53f8d72920ba2744fe873ebc.html HTTP/1.1" 404 177 "-" "Mozilla/5.0 (iPhone; CPU iPhone OS 6_0 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) Version/6.0 Mobile/10A5376e Safari/8536.25 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"`
	encoded = logging_internals.Encode_log_msg([]byte(example_nginx_log_line))
	util.Assert_result_equals_bytes(t, encoded, nil, example_nginx_log_line, 1)

	// log message containing newlines and unicode characters
	const input_with_newline = "hello \\\\ \nwo~~rl\n 世界 d!\n\n"
	const expected_result = `hello \x5c\x5c \x0awo~~rl\x0a \U00004e16\U0000754c d!\x0a\x0a` // look ma, no newlines! 'tis all escaped!
	encoded = logging_internals.Encode_log_msg([]byte(input_with_newline))
	util.Assert_result_equals_bytes(t, encoded, nil, expected_result, 1)
}

func Test_Decode_log_msg(t *testing.T) {
	t.Parallel()

	const original_str = `hello \\\\ \nw"@~;@#~;/"'£$%^&~~*(orl\n 世界 d!`
	encoded := logging_internals.Encode_log_msg([]byte(original_str))
	decoded := logging_internals.Decode_log_msg(encoded)

	util.Assert_result_equals_bytes(t, []byte(decoded), nil, original_str, 1)
}
