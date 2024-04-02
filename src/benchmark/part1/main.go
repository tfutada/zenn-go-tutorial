package main

import (
	"strconv"
)

func parseInput(s string) []byte {
	fish := make([]byte, 0, len(s)/2)

	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			break
		}
		if s[i] == ',' {
			continue
		}
		fish = append(fish, s[i])
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

// Solve function
func Solve(input string, days int) int {
	fish := parseInput(input)
	tempFish := []byte{}

	for i := 0; i < days; i++ {
		for j := range fish {
			if fish[j] == '0' {
				fish[j] = '6'
				tempFish = append(tempFish, '8')
			} else {
				fish[j]--
			}
		}

		fish = append(fish, tempFish...)
		tempFish = tempFish[:0]
	}

	return len(fish)
}

func main() {
	// call solve
	total := Solve("5,4,3,2,1", 10)
	println(total)
}
