package main

import (
	"strconv"
	"strings"
)

func parseInput(s string) []int {
	data := strings.Split(strings.Split(s, "\n")[0], ",")
	fish := make([]int, len(data))

	for i, d := range data {
		fish[i] = MustAtoi(d)
	}

	return fish
}

func MustAtoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return n
}

func Solve(input string, days int) int {
	fish := parseInput(input)

	for i := 0; i < days; i++ {
		tempFish := []int{}

		for j := range fish {
			if fish[j] == 0 {
				fish[j] = 6
				tempFish = append(tempFish, 8)
			} else {
				fish[j]--
			}
		}

		fish = append(fish, tempFish...)
	}

	return len(fish)
}

func main() {
	// call solve
	total := Solve("5,4,3,2,1", 10)
	println(total)
}
