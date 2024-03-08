package main

func main() {
	// Not Recommended
	a := make([]int, 10)
	a[0] = 1

	// Recommended
	b := make([]int, 0, 10)
	b = append(b, 1)

	// capacity of b
	println(cap(b))
}
