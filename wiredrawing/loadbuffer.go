package wiredrawing

import (
	"bufio"
	"fmt"
	"io"
)

var scanner *bufio.Scanner

var loadedBuffer string

var currentLine int

// ---------------------------------------------------------------------
// 引数に渡された io.ReadCloser 変数の中身を読み取り出力する
// ---------------------------------------------------------------------
func LoadBuffer(buffer io.ReadCloser, previousLine *int) (bool, error) {
	// fmt.Println("bufferの読み取り開始------>")
	currentLine = 0

	for {
		readData := make([]byte, 1)
		n, err := buffer.Read(readData)

		if err != nil {
			// fmt.Println(err)
			break
		}

		if n == 0 {
			break
		}
		if currentLine >= *previousLine {
			fmt.Print(string(readData))
		}
		currentLine++
		readData = nil
	}
	*previousLine = currentLine

	return true, nil
}
