package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	billionData()
}

func billionData() {
	// Simulated input data for the sake of example.
	data := `Yellowknife;16.0
Entebbe;32.9
Porto;24.4
Vilnius;12.4
Fresno;7.9
Maun;17.5
Panama City;39.5`

	// Splitting the input into lines.
	lines := strings.Split(data, "\n")

	// Variables to store the computation of min, max, and mean.
	var min, max, sum float64
	min = 1e9  // Set to a very high number initially.
	max = -1e9 // Set to a very low number initially.

	// Map to store the city and its temperature.
	temperatureData := make(map[string]float64)

	// Processing each line to parse the data.
	for _, line := range lines {
		parts := strings.Split(line, ";")
		if len(parts) != 2 {
			log.Fatal("Invalid data format")
		}
		city := parts[0]
		temp, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			log.Fatal(err)
		}

		// Storing data in map.
		temperatureData[city] = temp

		// Calculating min, max, and sum for mean.
		if temp < min {
			min = temp
		}
		if temp > max {
			max = temp
		}
		sum += temp
	}

	// Calculating the mean.
	mean := sum / float64(len(temperatureData))

	// Preparing to write to a file.
	file, err := os.Create("output.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Building the string to be written in the specified format.
	output := "{"
	for city, _ := range temperatureData {
		output += fmt.Sprintf("%s=%.1f/%.1f/%.1f, ", city, min, max, mean)
	}
	output = strings.TrimRight(output, ", ") + "}"

	// Writing to the file.
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(output)
	if err != nil {
		log.Fatal(err)
	}
	writer.Flush()

	fmt.Println("Data processed and written to file successfully.")
}
