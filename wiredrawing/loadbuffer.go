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

	// 前回の実行改行回数がnilなら0を代入
	if previousLine == nil {
		*previousLine = currentLine
	}
	scanner = bufio.NewScanner(buffer)
	for {
		// 読み取りが失敗した場合ループを抜ける
		if scanner.Scan() != true {
			// 読み取り失敗
			// fmt.Println("読み取り失敗")
			break
		}

		// 改行をカウントする
		currentLine++
		loadedBuffer = scanner.Text()
		if len(loadedBuffer) == 0 {
			continue
		}
		// 過去の出力内容は破棄する
		if *previousLine < currentLine {
			fmt.Println(loadedBuffer)
		}
		continue
	}
	// defer fmt.Println("bufferの読み取り完了------>")
	*previousLine = currentLine
	return true, nil
}
