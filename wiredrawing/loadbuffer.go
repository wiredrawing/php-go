package wiredrawing

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

var scanner *bufio.Scanner

var loadedBuffer string

var currentLine int

const ensureLength int = 512

// ---------------------------------------------------------------------
// 引数に渡された io.ReadCloser 変数の中身を読み取り出力する
// 2060
// ---------------------------------------------------------------------
func LoadBuffer(buffer io.ReadCloser, previousLine *int) (bool, error) {
	currentLine = 0

	for {
		readData := make([]byte, ensureLength)
		n, err := buffer.Read(readData)

		if err != nil {
			break
		}

		if n == 0 {
			break
		}

		from := currentLine
		to := currentLine + n
		if (currentLine + n) >= *previousLine {
			if from < *previousLine && *previousLine <= to {
				diff := *previousLine - currentLine
				tempSlice := readData[diff:]
				// fmt.Print(string(tempSlice))
				fmt.Fprint(os.Stdout, string(tempSlice))
			} else {
				// fmt.Print(string(readData))
				fmt.Fprint(os.Stdout, string(readData))
			}
		}
		currentLine += n
		readData = nil
	}
	*previousLine = currentLine

	return true, nil
}
