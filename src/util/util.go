package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func MustNewID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(fmt.Sprintf("failed to generate threadID: %s", err))
	}
	return base64.StdEncoding.EncodeToString(b)
}
