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
	//var readString []string
	var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
	//
	//var s = 0
	//for {
	//	if s > 0 {
	//		fmt.Print("  |> ")
	//	}
	//	scanner.Scan()
	//	value := scanner.Text()
	//
	//	if value == "rollback" {
	//		if len(readString) > 0 {
	//			os.Stdout.Write([]byte("\033[1A"))
	//			var lastString string = readString[len(readString)-1]
	//			readString = readString[0 : len(readString)-1]
	//			fmt.Print("\v  --- " + lastString + "\n")
	//			//fmt.Printf("rollback: %v\n len %v", readString, len(readString))
	//			continue
	//		}
	//	}
	//	if value == "" {
	//		break
	//	}
	//	readString = append(readString, value)
	//	s++
	//	//fmt.Printf("k: %v, v: %v\n", key, key)
	//}
	//if len(readString) > 0 {
	//	return strings.Join(readString, "\n")
	//}
	//return ""
	var ressult bool = scanner.Scan()
	if ressult == true {
		var textYouGet string = scanner.Text()
		return textYouGet
	}
	fmt.Println("Failed scanner.Scan().")
	return ""
}
