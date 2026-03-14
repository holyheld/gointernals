package rest

import (
	"encoding/json"
	"io"
)

func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func Encode(w io.Writer, data any) error {
	return json.NewEncoder(w).Encode(data)
}

func Decode(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}
