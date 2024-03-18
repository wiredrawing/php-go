package wiredrawing

import (
	"fmt"
	"io"
	"os"
	"runtime"
)

// LoadBuffer ---------------------------------------------------------------------
// 引数に渡された io.ReadCloser 変数の中身を読み取り出力する
// 2060
// ---------------------------------------------------------------------
func LoadBuffer(buffer io.ReadCloser, previousLine *int, showBuffer bool, whenError bool, colorCode string) (string, int) {
	var currentLine int

	const ensureLength int = 2048

	currentLine = 0
	var outputSize int = 0
	// whenError == true の場合バッファ内容を返却してやる
	var bufferWhenError string
	os.Stdout.WriteString("\033[" + colorCode + "m")
	for {
		readData := make([]byte, ensureLength)
		n, err := buffer.Read(readData)
		if (err != nil) && (err != io.EOF) {
			os.Stderr.Write([]byte(err.Error()))
			break
		}
		if n == 0 {
			break
		}
		readData = readData[:n]
		bufferWhenError += string(readData)

		from := currentLine
		to := currentLine + n
		if (currentLine + n) >= *previousLine {
			if from < *previousLine && *previousLine <= to {
				diff := *previousLine - currentLine
				tempSlice := readData[diff:]
				// 出力内容の表示フラグがtrueの場合のみ
				if showBuffer == true {
					outputSize += len(tempSlice)
					fmt.Fprint(os.Stdout, string(tempSlice))
				}
			} else {
				// 出力内容の表示フラグがtrueの場合のみ
				if showBuffer == true {
					outputSize += len(readData)
					fmt.Fprint(os.Stdout, string(readData))
				}
			}
		}
		currentLine += n
		readData = nil
	}
	// エラーチェック以外の場合
	if whenError != true {
		*previousLine = currentLine
	}
	// 使用したメモリを開放してみる
	runtime.GC()
	// コンソールのカラーをもとにもどす
	os.Stdout.WriteString("\033[0m")
	//debug.FreeOSMemory()

	return bufferWhenError, outputSize
}
