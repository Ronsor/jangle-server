package util

import (
	"log"
)

var NoPanic = true

func TryLog(i interface{}) {
	if i != nil {
		log.Println(i)
	}
}

func TryRecover() {
	if !NoPanic { return }
	TryLog(recover())
}
