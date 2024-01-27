package main

import "fmt"

	
func structCorruption() string {
	type Pair struct {
		X int
		Y int
	}

	arr := []Pair{{X: 0, Y: 0}, {X: 1, Y: 1}}
	var p Pair // 共有変数
	
	// writer
	go func() {
		for i := 0; ; i++ {
			// 代入するのは{X: 0, Y: 0}, {X: 1, Y: 1}のどちらかのみ
			p = arr[i%2] 
		}
	}()
	
	// reader
	for {
		read := p
		switch read.X + read.Y {
		case 0, 2: 
			// {X: 0, Y: 0}, {X: 1, Y: 1}のどちらかならば、
			// このケースに入るので何も起きない。
		default:
			return fmt.Sprintf("struct corruption detected: %+v", read)
		}
	}
}

func main() {
		fmt.Println(structCorruption())
}





