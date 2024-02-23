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
			ints, done := compare(nums, target, i, j)
			if done {
				return ints
			}
		}
	}
	// Return an empty slice if no solution is found.
	// Depending on the problem statement, you may need to handle this case differently.
	return []int{}
}

func compare(nums []int, target int, i int, j int) ([]int, bool) {
	if nums[i]+nums[j] == target {
		return []int{i, j}, true
	}
	return nil, false
}
