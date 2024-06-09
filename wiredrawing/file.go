package wiredrawing

import (
	"os"
)

// FileOpen -------------------------------------------------------------
// 指定したファイルパスのファイルを開き
// 第二引数の文字列を書き込む処理
// -------------------------------------------------------------
func FileOpen(filePath string, text string) (int, error) {
	var file *os.File = nil
	var err error = nil
	file, err = os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0777)
	// 本関数は実行の度にファイルを開き,都度閉じる
	defer (func() {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	})()
	if err != nil {
		return 0, err
	}
	var bytesWritten int

	bytesWritten, err = file.WriteString(text)

	if err != nil {
		return bytesWritten, err
	}

	return bytesWritten, err
}
