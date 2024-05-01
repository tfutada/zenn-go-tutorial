package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type StationData struct {
	Name  string
	Min   float64
	Max   float64
	Sum   float64
	Count int
}

func run() {
	// Load the station data from a file
	file, err := os.Open("measurements.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	data := make(map[string]*StationData)
	scanner := bufio.NewScanner(file)

	// Parse each line
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ";")
		if len(parts) != 2 {
			continue // skip malformed lines
		}
		name := parts[0]
		tempStr := strings.TrimSpace(parts[1])

		temperature, err := strconv.ParseFloat(tempStr, 64)
		if err != nil {
			panic(err)
		}

		if station, exists := data[name]; !exists {
			data[name] = &StationData{name, temperature, temperature, temperature, 1}
		} else {
			if temperature < station.Min {
				station.Min = temperature
			}
			if temperature > station.Max {
				station.Max = temperature
			}
			station.Sum += temperature
			station.Count++
		}
	}

	printResult(data)
}

func printResult(data map[string]*StationData) {
	// Prepare data for output
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fmt.Print("{")
	for i, key := range keys {
		station := data[key]
		fmt.Printf("%s=%.1f/%.1f/%.1f", key, station.Min, station.Sum/float64(station.Count), station.Max)
		if i < len(keys)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Println("}")
}

func main() {
	started := time.Now()
	run()
	fmt.Printf("Execution time: %0.6f seconds\n", time.Since(started).Seconds())
}
