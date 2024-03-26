package wiredrawing

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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

	// 入力モードの選択用入力
	var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
	var ressult bool = scanner.Scan()
	if ressult == true {
		var which string = scanner.Text()
		if which == ">>>" {
			fmt.Print("\033[33m")
			var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
			var readString []string
			var s = 0
			for {
				if s > 0 {
					fmt.Printf("%s%s", " ... ", " ... ")
				} else {
					fmt.Printf("%s%s", " ... ", " >>> ")
				}
				scanner.Scan()
				value := scanner.Text()

				if value == "rollback" {
					if len(readString) > 0 {
						os.Stdout.Write([]byte("\033[1A"))
						var lastString string = readString[len(readString)-1]
						readString = readString[0 : len(readString)-1]
						fmt.Print("\v  --- " + lastString + "\n")
						//fmt.Printf("rollback: %v\n len %v", readString, len(readString))
						continue
					}
				} else if value == "cat" {
					// 現在までの入力を確認する
					var indexStr string = ""
					for index, value := range readString {
						indexStr = fmt.Sprintf("%03d", index)
						fmt.Print(colorWrapping("34", indexStr) + ": ")
						fmt.Println(colorWrapping("32", value))
					}
					continue
				}
				if value == "" {
					break
				}
				readString = append(readString, value)
				s++
				//fmt.Printf("k: %v, v: %v\n", key, key)
			}
			fmt.Print("\033[0m")
			if len(readString) > 0 {
				return strings.Join(readString, "\n")
			}
			return ""
		} else {
			return which
		}
	}
	//fmt.Println("Failed scanner.Scan().")
	return ""
}

func colorWrapping(colorCode string, text string) string {
	return "\033[" + colorCode + "m" + text + "\033[0m"
}
