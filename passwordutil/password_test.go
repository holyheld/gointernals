package passwordutil_test

import (
	"testing"

	"github.com/holyheld/passwordutil"
)

func TestPasswordHashing(t *testing.T) {
	t.Parallel()

	pass := "1234"

	hash, err := passwordutil.GeneratePasswordHash(pass)
	if err != nil {
		t.Fatalf("GeneratePasswordHash return err: %s", err)

		return
	}

	_, err = passwordutil.ComparePasswordAndHash(pass, hash)
	if err != nil {
		t.Fatalf("ComparePasswordAndHash return err: %s", err)

		return
	}
}
