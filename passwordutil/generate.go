package passwordutil

import (
	"crypto/rand"
	"strings"
)

// Generate returns random password of specified length
//
// Note that recommended length is 26, based on [rand.Text] recommendation,
// to get 128 bits of randomness, which is plenty for bootstrap password.
func Generate(length int) string {
	sb := strings.Builder{}

	for i := 0; i < length; i += 26 {
		sb.WriteString(rand.Text())
	}

	return sb.String()[0:length]
}
