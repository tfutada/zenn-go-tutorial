package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
)

func main() {

	// Create a file to store the CPU profile.
	cpuProfile, err := os.Create("cpu.pprof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	// Ensure that the profile will be closed and written when the function returns.
	defer cpuProfile.Close()

	// Start the CPU profiler.
	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	// Ensure that the profiler is stopped before the function returns.
	defer pprof.StopCPUProfile()

	// main func
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
