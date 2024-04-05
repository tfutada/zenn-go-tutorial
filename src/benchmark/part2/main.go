package main

// Part2 of https://dev.to/gkampitakis/advent-of-code-investigating-performance-improvements-in-go-2l07
func Solve(input string, days int) int {
	fish := parseInput(input)
	tmpFish := make(map[uint8]int, 9)

	for i := 0; i < days; i++ {
		resetCount := 0

		for k, v := range fish {
			if k == 0 {
				resetCount += v
			} else {
				tmpFish[k-1] = v
			}
		}

		tmpFish[8] += resetCount
		tmpFish[6] += resetCount

		fish, tmpFish = tmpFish, fish
		clear(tmpFish)
	}

	counter := 0
	for _, v := range fish {
		counter += v
	}

	return counter
}

func parseInput(s string) map[uint8]int {
	fish := make(map[uint8]int, 9)

	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			break
		}
		if s[i] == ',' {
			continue
		}
		fish[s[i]-'0']++
	}

	return fish
}

func main() {
	// call solve
	total := Solve("5,4,3,2,1", 10)
	println(total)
}
