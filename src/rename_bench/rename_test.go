package main

import (
	"os"
	"testing"
)

func renameFile(oldName, newName string) error {
	return os.Rename(oldName, newName)
}

func writeDummyData(fileName string, size int) error {
	data := make([]byte, size)
	for i := range data {
		data[i] = 'A' // Writing 'A' character to fill the file
	}
	return os.WriteFile(fileName, data, 0644)
}

func readFileToMemory(fileName string) ([]byte, error) {
	return os.ReadFile(fileName)
}

func BenchmarkRenameAndReadFile(b *testing.B) {
	oldName := "oldfile.txt"
	newName := "newfile.txt"
	dummyDataSize := 10 * 1024 // 10KB

	for i := 0; i < b.N; i++ {
		// Create and write dummy data to the file
		err := writeDummyData(oldName, dummyDataSize)
		if err != nil {
			b.Fatalf("Failed to write dummy data: %v", err)
		}

		// Rename the file
		err = renameFile(oldName, newName)
		if err != nil {
			b.Fatalf("Error renaming file: %v", err)
		}

		// Read the renamed file into memory
		_, err = readFileToMemory(newName)
		if err != nil {
			b.Fatalf("Error reading file: %v", err)
		}

		// Clean up by removing the file
		err = os.Remove(newName)
		if err != nil {
			b.Fatalf("Error removing file: %v", err)
		}
	}
}

func TestRenameAndReadFile(t *testing.T) {
	oldName := "oldfile.txt"
	newName := "newfile.txt"
	dummyDataSize := 10 * 1024 // 10KB

	// Create and write dummy data to the file
	err := writeDummyData(oldName, dummyDataSize)
	if err != nil {
		t.Fatalf("Failed to write dummy data: %v", err)
	}

	// Test renaming the file
	err = renameFile(oldName, newName)
	if err != nil {
		t.Fatalf("Error renaming file: %v", err)
	}

	// Read the file into memory
	_, err = readFileToMemory(newName)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	// Clean up by removing the file
	err = os.Remove(newName)
	if err != nil {
		t.Fatalf("Error removing file: %v", err)
	}
}
