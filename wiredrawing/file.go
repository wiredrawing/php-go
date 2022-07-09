package wiredrawing

import (
	"os"
)

var file *os.File = nil
var err error = nil

// FileOpen -------------------------------------------------------------
// 指定したファイルパスのファイルを開き
// 第二引数の文字列を書き込む処理
// -------------------------------------------------------------
func FileOpen(filePath string, text string) (int, error) {

	file, err = os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0777)

	if err != nil {
		return 0, err
	}
	var bytesWritten int

	bytesWritten, err = file.WriteString(text)

	if err != nil {
		return bytesWritten, err
	}

	// 本関数は実行の度にファイルを開き,都度閉じる
	defer (func() {
		file.Close()
	})()

	return bytesWritten, err
}
