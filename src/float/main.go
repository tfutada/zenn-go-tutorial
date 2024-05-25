package main

import (
	"fmt"
	"math"
)

func main() {
	// Print floating-point numbers with high precision
	fmt.Printf("%.55f\n", 0.1)
	fmt.Printf("%.55f\n", 0.2)
	fmt.Printf("%.55f\n", 0.3)

	fmt.Printf("%.55f\n", 0.1+0.2)

	// Comparing floating-point numbers directly
	if 0.1+0.2 == 0.3 {
		fmt.Println("0.1 + 0.2 is equal to 0.3")
	} else {
		fmt.Println("0.1 + 0.2 is not equal to 0.3")
	}

	// Using math/big for floating-point comparison
	if isClose(0.1+0.2, 0.3, 1e-9) {
		fmt.Println("0.1 + 0.2 is approximately equal to 0.3")
	} else {
		fmt.Println("0.1 + 0.2 is not approximately equal to 0.3")
	}

	// Another example of floating-point precision
	fmt.Printf("%.55f\n", 100*1.1)
}

// isClose checks if two float64 values are approximately equal within a tolerance
func isClose(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
