package main

import (
	// 標準パッケージ
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"sync"

	// ここは独自パッケージ
	"go-sample/samplepackage"
	"go-sample/wiredrawing"

	// _をつけた場合は パッケージ内のinit関数のみ実行される
	_ "go-sample/wiredrawing"

	"go-sample/wiredrawing2"

	"github.com/spf13/cobra"
	"golang.org/x/sys/windows"
	"rsc.io/quote"
)

var command *cobra.Command = new(cobra.Command)

// 割り込み監視用
var signal_chan chan os.Signal = make(chan os.Signal)

func main() {

	// コンソールの監視
	signal.Notify(
		signal_chan,
		os.Interrupt,
		os.Kill,
		windows.SIGKILL,
		windows.SIGHUP,
		windows.SIGINT,
		windows.SIGTERM,
		windows.SIGQUIT,
		windows.Signal(0x13),
		windows.Signal(0x14), // Windowsの場合 SIGTSTPを認識しないためリテラルで指定する
	)
	// ----------------------------------------------
	// 標準入力を可能にする
	// 標準入力の開始
	// ----------------------------------------------
	scanner := bufio.NewScanner(os.Stdin)

	var inputText string = ""
	for {
		fmt.Print(" > ")
		var isOk bool = scanner.Scan()
		if isOk != true {
			break
		}
		inputText = scanner.Text()
		fmt.Println(inputText)
	}
	// 標準入力の終了

	// cobraコマンドの初期化
	command.Use = "使い方"
	command.Short = "some descritpion"
	command.Long = "some long description"
	command.Run = func(cmd *cobra.Command, arguments []string) {
		fmt.Println(arguments)
	}

	// cobraの実行
	var err error = command.Execute()
	if err != nil {
		fmt.Print("Some Error Happend")
		panic(err)
		os.Exit(-1)
	}
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

	// goroutineのテスト
	var _waiter *sync.WaitGroup = new(sync.WaitGroup)
	echo := fmt.Println
	_waiter.Add(1)
	go (func(waiter *sync.WaitGroup) {
		// waiter.Add(1)
		echo("これはGoroutineの実行中です")
		defer waiter.Done()
	})(_waiter)
	// time.Sleep(10)
	_waiter.Wait()
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
