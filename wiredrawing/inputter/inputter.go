package inputter

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

// 入力用ポインタ
var scanner *bufio.Scanner

// 標準入力を待ち受ける関数
func StandByInput(waiter *sync.WaitGroup) {

	// ----------------------------------------------
	// 標準入力を可能にする
	// 標準入力の開始
	// ----------------------------------------------
	scanner = bufio.NewScanner(os.Stdin)

	var inputText string = ""
	for {
		fmt.Print("  >> ")
		var isOk bool = scanner.Scan()
		if isOk != true {
			fmt.Println("scanner.Scan()が失敗")
			// scannerを初期化
			scanner = nil
			scanner = bufio.NewScanner(os.Stdin)
			continue
		}
		inputText = scanner.Text()
		// 入力内容が exit ならアプリケーションを終了
		if len(inputText) > 0 {
			if inputText == "exit" {
				waiter.Done()
				// os.Exit(1)
			}
			fmt.Print(" ==> ")
			fmt.Println(inputText)
		}
	}
	// 標準入力の終了

}
