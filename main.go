package main

import (
	// 標準パッケージ
	"fmt"
	"sync"

	// ここは独自パッケージ
	"go-sample/wiredrawing"
	_ "go-sample/wiredrawing"

	"go-sample/wiredrawing2"

	"rsc.io/quote"
)

func main() {

	// 外部パッケージの構造体のポインタ変数を作成する
	var article *wiredrawing2.Article = new(wiredrawing2.Article)
	fmt.Println(article)
	wiredrawing.Print()
	fmt.Println("test")
	var wg sync.WaitGroup
	fmt.Println(wg)
	fmt.Println(quote.Hello())

}
