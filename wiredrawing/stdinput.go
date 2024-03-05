package wiredrawing

import (
	"bufio"
	"fmt"
	"os"
)

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
	fmt.Fprintln(os.Stdout, "Failed scanner.Scan().")
	return ""
}
