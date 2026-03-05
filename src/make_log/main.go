package main

import (
    "os"
)

func main() {
    fileSize := 1024 * 1024 // 1MB
    bufferSize := 1024      // 1KB buffer
    buffer := make([]byte, bufferSize)

    // Creating a file
    file, err := os.Create("big.log")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // Writing to the file
    for bytesWritten := 0; bytesWritten < fileSize; bytesWritten += bufferSize {
        if fileSize-bytesWritten < bufferSize {
            buffer = buffer[:fileSize-bytesWritten]
        }
        _, err := file.Write(buffer)
        if err != nil {
            panic(err)
        }
    }
}

