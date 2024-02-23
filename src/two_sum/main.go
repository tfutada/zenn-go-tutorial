package main

import "fmt"

func main() {
	nums := []int{2, 7, 11, 15}
	target := 18
	result := TwoSumBruteForce(nums, target)
	fmt.Println("Indices:", result)
}

// TwoSumBruteForce returns two values, whose sum is equals to target.
func TwoSumBruteForce(nums []int, target int) []int {
	for i := 0; i < len(nums); i++ {
		for j := i + 1; j < len(nums); j++ {
			if nums[i]+nums[j] == target {
				return []int{i, j}
			}
		}
	}
	// Return an empty slice if no solution is found.
	// Depending on the problem statement, you may need to handle this case differently.
	return []int{}
}
