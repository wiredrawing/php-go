package main

import (
	// 標準パッケージ
	"fmt"
	"sync"

	// ここは独自パッケージ
	"go-sample/samplepackage"
	"go-sample/wiredrawing"
	_ "go-sample/wiredrawing"

	"go-sample/wiredrawing2"

	"rsc.io/quote"
)

func main() {

	// 外部パッケージの構造体のポインタ変数を作成する
	var article *wiredrawing2.Article = new(wiredrawing2.Article)
	article.SetTitle("set the title")
	article.SetDescription("set the description")
	fmt.Println(article)
	wiredrawing.Print()
	fmt.Println("test")
	var wg sync.WaitGroup
	fmt.Println(wg)
	fmt.Println(quote.Hello())

	// execute concurrency
	var wg2 *sync.WaitGroup = new(sync.WaitGroup)
	var result string = functionForConcurrency(wg2)
	fmt.Println(result)
	wiredrawing.Print()
	samplepackage.CallableFunctionFromOtherPackage()
}

// --------------------------------------
// 並行処理で実行するための関数
// --------------------------------------
func functionForConcurrency(waiter *sync.WaitGroup) string {
	waiter.Add(1)
	// waitGroupをカウントダウンさせる
	defer waiter.Done()
	return "Return the some data you want to back"
}
