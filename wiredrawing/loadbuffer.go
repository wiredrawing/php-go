package wiredrawing

import (
	"bufio"
	"fmt"
	"io"
)

var scanner *bufio.Scanner

var loadedBuffer string

// 引数に渡された io.ReadCloser 変数の中身を読み取り出力する
func LoadBuffer(buffer io.ReadCloser) (bool, error) {
	fmt.Println("bufferの読み取り開始------>")
	scanner = bufio.NewScanner(buffer)
	for {
		if scanner.Scan() == true {
			loadedBuffer = scanner.Text()
			if len(loadedBuffer) == 0 {
				continue
			}
			fmt.Println(loadedBuffer)
			continue
		}
		// 読み取り失敗
		// fmt.Println("読み取り失敗")
		break
	}
	defer fmt.Println("bufferの読み取り完了------>")
	return true, nil
}
