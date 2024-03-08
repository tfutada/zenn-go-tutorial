package lab

import (
	"fmt"
	"log"
)

func init() {
	fmt.Println("init called from underscore package")
	log.SetPrefix("LOG: ")
}
