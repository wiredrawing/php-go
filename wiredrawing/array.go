package wiredrawing

import (
	"bufio"
	"fmt"
	"os"
)

// InArray ------------------------------------------------
// PHPのin_array関数をシミュレーション
// 第二引数 haystackに第一引数 needleが含まれていれば
// true それ以外は false
// ------------------------------------------------
func InArray(needle string, haystack []string) bool {

	// 第二引数に指定されたスライスをループさせる
	for _, value := range haystack {
		if needle == value {
			return true
		}
	}

	return false
}

// ArraySearch ------------------------------------------------
// PHPのarray_search関数をシミュレーション
// 第一引数にマッチする要素のキーを返却
// 要素が対象のスライス内に存在しない場合は-1
// ------------------------------------------------
func ArraySearch(needle string, haystack []string) int {

	for index, value := range haystack {
		if value == needle {
			return index
		}
	}
	return -1
}

// StdInput ----------------------------------------
// 標準入力から入力された内容を文字列で返却する
// ----------------------------------------
func StdInput() string {

	var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)

	var ressult bool = scanner.Scan()
	if ressult == true {
		var textYouGet string = scanner.Text()
		return textYouGet
	}
	fmt.Println("Failed scanner.Scan().")
	return ""
}
