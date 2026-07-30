package typeutil_test

import (
	"encoding/json"

	"github.com/holyheld/gointernals/typeutil"
)

func ExampleMust() {
	data := map[string]string{
		"hello": "world",
	}

	typeutil.Must(json.Marshal(data))
	// Output: []byte(`{"hello":"world"}`)
}
