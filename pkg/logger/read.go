package logger

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func ReadLastNLines(path string, n int) ([]string, error) {
	// Open the file
	file, err := os.Open(fmt.Sprintf("%s/%s", path, accessFilename))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fileInfo.Size()

	// If file is empty, return empty slice
	if fileSize == 0 {
		return []string{}, nil
	}

	// Buffer for reading
	bufferSize := int64(4096)
	if bufferSize > fileSize {
		bufferSize = fileSize
	}
	buffer := make([]byte, bufferSize)

	// Start reading from the end
	position := fileSize
	lines := make([]string, 0, n)
	lineCount := 0

	for lineCount < n && position > 0 {
		// How much to read
		readSize := bufferSize
		if position < bufferSize {
			readSize = position
		}
		position -= readSize

		// Read chunk from position
		_, err := file.Seek(position, io.SeekStart)
		if err != nil {
			return nil, err
		}

		_, err = file.Read(buffer[:readSize])
		if err != nil {
			return nil, err
		}

		// Count newlines in reverse
		for i := readSize - 1; i >= 0; i-- {
			if buffer[i] == '\n' {
				lineCount++
				if lineCount > n {
					// We found more than n lines
					// Need to adjust position to read only last n lines
					position += int64(i) + 1
					break
				}
			}
		}
	}

	// If we couldn't find n lines, start from beginning
	if position < 0 {
		position = 0
	}

	// Seek to the position where we want to start reading
	_, err = file.Seek(position, io.SeekStart)
	if err != nil {
		return nil, err
	}

	// Read lines from position to end
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Check if we need to trim
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	return lines, scanner.Err()
}
