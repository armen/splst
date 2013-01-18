package utils

import (
	"crypto/rand"
	"fmt"
	"io"
)

func GenId(length uint) string {

	buf := make([]byte, length)
	io.ReadFull(rand.Reader, buf)

	return fmt.Sprintf("%x", buf)
}
