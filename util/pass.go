package util

import (
	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
)

func CryptPass(pwd string) string {
	cr := crypt.SHA512.New()
	ret, _ := cr.Generate([]byte(pwd), []byte{})
	return ret
}

func VerifyPass(hash string, pwd string) bool {
	cr := crypt.SHA512.New()
	err := cr.Verify(hash, []byte(pwd))
	return err == nil
}
