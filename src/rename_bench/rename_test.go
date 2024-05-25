package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func renameFile(oldName, newName string) error {
	return os.Rename(oldName, newName)
}

func readFileToMemory(fileName string) ([]byte, error) {
	return ioutil.ReadFile(fileName)
}

func BenchmarkRenameAndReadFile(b *testing.B) {
	oldName := "oldfile.txt"
	newName := "newfile.txt"

	for i := 0; i < b.N; i++ {
		// Create a dummy file to rename
		file, err := os.Create(oldName)
		if err != nil {
			b.Fatalf("Failed to create file: %v", err)
		}
		file.Close()

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

	// Create a dummy file to rename
	file, err := os.Create(oldName)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	file.Close()

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
