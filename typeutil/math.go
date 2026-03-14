package typeutil

import (
	"math"
)

type numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Ceil returns the least integer value greater than or equal to x, casted to type T.
func Ceil[T numeric](x float64) T {
	return T(math.Ceil(x))
}

// DivUp divides x, y (converted to float64 for accuracy first), then applies [Ceil]
// on top.
func DivUp[T numeric](x, y T) T {
	return Ceil[T](float64(x) / float64(y))
}
