package utils

import (
	"crypto/rand"
	"fmt"
	"io"
)

func GenId() string {
	buf := make([]byte, 16)
	io.ReadFull(rand.Reader, buf)
	return fmt.Sprintf("%x", buf)
}
