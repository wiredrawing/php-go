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
// 2060
// ---------------------------------------------------------------------
func LoadBuffer(buffer io.ReadCloser, previousLine *int) (bool, error) {
	// fmt.Println("bufferの読み取り開始------>")
	currentLine = 0

	var ensureLength = 1024
	fmt.Println("*previousLine", *previousLine)
	for {
		readData := make([]byte, ensureLength)
		n, err := buffer.Read(readData)

		if err != nil {
			break
		}

		if n == 0 {
			break
		}

		if (currentLine + n) >= *previousLine {

			if currentLine < *previousLine && *previousLine < (currentLine+n) {
				fmt.Print(string(readData[(*previousLine - currentLine):]))
			} else {
				fmt.Print(string(readData))
			}
		}
		// if currentLine >= *previousLine {
		// 	//2048 2060
		// 	fmt.Print(string(readData))
		// }
		currentLine += n
		readData = nil
	}
	*previousLine = currentLine

	// fmt.Println("*previousLine", *previousLine)
	// fmt.Println("currentLine", currentLine)
	return true, nil
}
