// add_test.go
package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestTwoSumBruteForce
func TestTwoSumBruteForce(t *testing.T) {
	// Define test cases
	testCases := []struct {
		nums     []int
		target   int
		expected []int // Adding an expected slice to validate against
	}{
		{nums: []int{2, 7, 11, 15}, target: 9, expected: []int{0, 1}},
		{nums: []int{3, 2, 4}, target: 6, expected: []int{1, 2}},
		{nums: []int{-1, -2, -3, -4, -5}, target: -8, expected: []int{2, 4}},
		{nums: []int{1, 5, 3, 7, 9}, target: 12, expected: []int{1, 3}},
		{nums: []int{0, 4, 3, 0}, target: 0, expected: []int{0, 3}},
		{nums: []int{1, 2, 3, 4, 5, 6}, target: 11, expected: []int{4, 5}},
		// Test case where no two numbers sum up to the target
		{nums: []int{1, 2, 3, 9}, target: 17, expected: []int{}},
		// Test case with duplicates in the array
		{nums: []int{1, 1, 3, 4, 5}, target: 2, expected: []int{0, 1}},
		// Test case with target as zero and negative numbers
		{nums: []int{-3, 4, 3, 90}, target: 0, expected: []int{0, 2}},
	}

	// Run test cases
	for _, tc := range testCases {
		result := TwoSumBruteForce(tc.nums, tc.target)
		assert.Equal(t, tc.expected, result)
	}
}

// BenchmarkAdd benchmarks the Add function.
func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		TwoSumBruteForce([]int{2, 7, 11, 15}, 9)
	}
}
