package rest

import (
	"encoding/json"
	"io"
)

func Marshal(v any) ([]byte, error) {
	return MarshalCustom(defaultSerializer, v)
}

func Unmarshal(data []byte, v any) error {
	return UnmarshalCustom(defaultSerializer, data, v)
}

func Encode(w io.Writer, data any) error {
	return EncodeCustom(defaultSerializer, w, data)
}

func Decode(r io.Reader, v any) error {
	return DecodeCustom(defaultSerializer, r, v)
}

func MarshalCustom(m Marshaler, v any) ([]byte, error) {
	return m.Marshal(v)
}

func UnmarshalCustom(u Unmarshaler, data []byte, v any) error {
	return u.Unmarshal(data, v)
}

func EncodeCustom(e Encoder, w io.Writer, data any) error {
	return e.Encode(w, data)
}

func DecodeCustom(d Decoder, r io.Reader, v any) error {
	return d.Decode(r, v)
}

type Marshaler interface {
	Marshal(data any) ([]byte, error)
}

type Unmarshaler interface {
	Unmarshal(data []byte, dst any) error
}

type Encoder interface {
	Encode(w io.Writer, src any) error
}

type Decoder interface {
	Decode(r io.Reader, dst any) error
}

type Serializer interface {
	Marshaler
	Unmarshaler
	Encoder
	Decoder
}

var defaultSerializer Serializer = &jsonSerializer{}

func SetDefaultSerializer(s Serializer) {
	defaultSerializer = s
}

func DefaultSerializer() Serializer {
	return defaultSerializer
}

type jsonSerializer struct{}

func (s *jsonSerializer) Marshal(data any) ([]byte, error) {
	return json.Marshal(data)
}

func (s *jsonSerializer) Unmarshal(data []byte, dst any) error {
	return json.Unmarshal(data, dst)
}

func (s *jsonSerializer) Encode(w io.Writer, src any) error {
	return json.NewEncoder(w).Encode(src)
}

func (s *jsonSerializer) Decode(r io.Reader, dst any) error {
	return json.NewDecoder(r).Decode(dst)
}
