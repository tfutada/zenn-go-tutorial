package main

import (
	"io"
)


func main()  {
	
}

// サイズがFsizeのファイルをnbyteごと読む関数
func ReadOS(r io.Reader, n int, Fsize int) {
	data := make([]byte, n)

	t := Fsize / n
	for i := 0; i < t; i++ {
		r.Read(data)
	}
}

