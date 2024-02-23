package main

import (
	"io"
)

func main() {

}

// ReadOS サイズがFsizeのファイルをnbyteごと読む関数
func ReadOS(r io.Reader, n int, Fsize int) {
	data := make([]byte, n)

	t := Fsize / n
	for i := 0; i < t; i++ {
		r.Read(data)
	}
}
