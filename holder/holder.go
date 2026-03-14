package holder

type Holder[T any] interface {
	Get() T
}
