package jsonv1

import (
	"bytes"
	"encoding/json"
)

func encode(v interface{}, escapeHTML bool, indent string) *bytes.Buffer {
	var buffer = new(bytes.Buffer)
	var enc = json.NewEncoder(buffer)
	enc.SetEscapeHTML(escapeHTML)
	enc.SetIndent("", indent)
	var _ = enc.Encode(v)
	return buffer
}

// IndentEncode encode go interface to buffer with indent
func IndentEncode(v interface{}) *bytes.Buffer {
	return encode(v, false, "  ")
}

// Encode go interface{} to buffer
func Encode(v interface{}) *bytes.Buffer {
	return encode(v, false, "")
}

// Decode buffer to go interface{}
func Decode(bf *bytes.Buffer, v interface{}) error {
	var dec = json.NewDecoder(bf)
	return dec.Decode(v)
}
