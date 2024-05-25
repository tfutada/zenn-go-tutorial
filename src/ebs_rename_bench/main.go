package main

import (
	"fmt"
	"os"
)

func renameFile(oldName, newName string) error {
	return os.Rename(oldName, newName)
}

func main() {
	oldName := "oldfile.txt"
	newName := "newfile.txt"

	err := renameFile(oldName, newName)
	if err != nil {
		fmt.Println("Error renaming file:", err)
		return
	}

	fmt.Println("File renamed successfully")
}
