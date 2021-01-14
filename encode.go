package gostomp

import (
	"bytes"
	"strings"
)

var (
	replacerEncodeValues = strings.NewReplacer(
		"\\", "\\\\",
		"\r", "\\r",
		"\n", "\\n",
		":", "\\c",
	)
	replacerDecodeValues = strings.NewReplacer(
		"\\r", "\r",
		"\\n", "\n",
		"\\c", ":",
		"\\\\", "\\",
	)
)

// Encodes a header value using STOMP value encoding
func encodeValue(s string) []byte {
	var buf bytes.Buffer
	buf.Grow(len(s))
	replacerEncodeValues.WriteString(&buf, s)
	return buf.Bytes()
}

// Unencodes a header value using STOMP value encoding
func decodeValue(b []byte) (string, error) {
	s := replacerDecodeValues.Replace(string(b))
	return s, nil
}
