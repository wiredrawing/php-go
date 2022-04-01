package main

import (
	// 標準パッケージ

	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"time"

	// ここは独自パッケージ

	// _をつけた場合は パッケージ内のinit関数のみ実行される
	"go-sample/wiredrawing"
	_ "go-sample/wiredrawing"

	"go-sample/wiredrawing/inputter"
	"go-sample/wiredrawing/parallel"

	"golang.org/x/sys/windows"
)

// var command *cobra.Command = new(cobra.Command)

// 割り込み監視用
var signal_chan chan os.Signal = make(chan os.Signal)

// ガベージコレクションを任意の時間間隔で実行
func regularsGabageCollection() {

	for {
		time.Sleep(5 * time.Second)
		runtime.GC()
	}
}

var previousLine *int = new(int)

func main() {

	fmt.Println("0")
	*previousLine = 0

	// コマンドの実行結果をPipeで受け取る
	command := exec.Command("php", "-i")
	buffer, err := command.StdoutPipe()

	if err != nil {
		panic(err)
	}
	command.Start()

	fmt.Println("1")
	wiredrawing.LoadBuffer(buffer, previousLine)
	fmt.Println("2")
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

	// GCを実行
	go regularsGabageCollection()

	var exit chan int = make(chan int)
	// 割り込み対処を実行するGoルーチン
	go parallel.InterruptProcess(exit, signal_chan)

	go func(exit chan int) {
		var echo = fmt.Print
		var code int = 0
		for {
			code, _ = <-exit

			if code == 1 {
				os.Exit(code)
			} else if code == 4 {
				echo("[Ignored interrupt].\r\n")
			} else {
				if runtime.GOOS != "darwin" {
					echo("[Ignored interrupt].\r\n")
				}
			}
		}
	}(exit)

	// ----------------------------------------------
	// 標準入力を可能にする
	// 標準入力の開始
	// ----------------------------------------------

	var waiter *sync.WaitGroup = new(sync.WaitGroup)
	waiter.Add(1)
	go inputter.StandByInput(waiter)
	waiter.Wait()

}
