package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	stringPointer5()
}

func stringPointer() {
	var s *string
	tmp := "hello"
	s = &tmp
	fmt.Printf("address: %+v, value: %s", s, *s)
}

//func stringPointer2() {
//	tmp := "hello"
//	tmp[0] = 'J'
//	fmt.Println(tmp)
//}

func stringPointer3() {
	tmp := "hello"
	tmp_str := []byte(tmp)
	tmp_str[0] = 'J'
	fmt.Println(string(tmp_str))

	tmp2 := "😊orld" // contains a 4-byte character
	tmp_str2 := []byte(tmp2)
	tmp_str2[0] = 'J' // this will corrupt the UTF-8 encoding
	fmt.Println(string(tmp_str2))
}

func stringPointer4() {
	tmp := "€"
	fmt.Println("bytes: ", len(tmp))                    // prints: 3
	fmt.Println("runes: ", utf8.RuneCountInString(tmp)) // prints: 1
}

func stringPointer5() {
	tmp := "❤€%……&*"
	fmt.Printf("char at 0 index, has type %T and value is %+v\n", tmp[0], tmp[0])

	for _, t := range tmp {
		fmt.Printf("value is %+v type is %T\n", t, t)

	}
}
