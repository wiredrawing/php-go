package wiredrawing

import (
	"bufio"
	"fmt"
	"os"
)

// ----------------------------------------
// 標準入力から入力された内容を文字列で返却する
// ----------------------------------------
func StdInput() string {

	var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)

	if scanner.Scan() {
		var textYouGet string = scanner.Text()
		return textYouGet
	}
	fmt.Fprintln(os.Stdout, "Failed scanner.Scan().")
	return ""
}
