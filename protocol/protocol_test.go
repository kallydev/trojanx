package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/google/uuid"
	"testing"
)

func Test_sha224(t *testing.T) {
	hash := sha256.New224()
	password := uuid.NewString()
	hash.Write([]byte(password))
	t.Log(password, hex.EncodeToString(hash.Sum(nil)))
}
