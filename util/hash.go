package util

import (
	"fmt"
	"crypto/md5"
)

func MD5(in []byte) string {
	return fmt.Sprintf("%x", md5.Sum(in))
}
