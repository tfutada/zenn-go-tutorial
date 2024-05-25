package main

import (
	"os"
	"testing"
)

func BenchmarkRenameFile(b *testing.B) {
	oldName := "oldfile.txt"
	newName := "newfile.txt"

	// Create a dummy file to rename
	file, err := os.Create(oldName)
	if err != nil {
		b.Fatalf("Failed to create file: %v", err)
	}
	file.Close()

	// Reset the benchmark timer to exclude setup time
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Rename the file
		err := renameFile(oldName, newName)
		if err != nil {
			b.Fatalf("Error renaming file: %v", err)
		}

		// Rename it back to the original name for the next iteration
		err = renameFile(newName, oldName)
		if err != nil {
			b.Fatalf("Error renaming file back: %v", err)
		}
	}
}

func TestRenameFile(t *testing.T) {
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

	// Clean up by renaming it back to the original name
	err = renameFile(newName, oldName)
	if err != nil {
		t.Fatalf("Error renaming file back: %v", err)
	}
}
