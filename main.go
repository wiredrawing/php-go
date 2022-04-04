package main

import (
	// 標準パッケージ

	"fmt"
	"os"
	"os/signal"
	"runtime"

	// ここは独自パッケージ

	// _をつけた場合は パッケージ内のinit関数のみ実行される

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

	// var mem runtime.MemStats
	// for {

	// 	runtime.ReadMemStats(&mem)
	// 	// fmt.Printf("(1)Alloc:%d, (2)TotalAlloc:%d, (3)Sys:%d, (4)HeapAlloc:%d, (5)HeapSys:%d, (6)HeapReleased:%d\r\n",
	// 	// 	mem.Alloc, // HeapAllocと同値
	// 	// 	mem.TotalAlloc,
	// 	// 	mem.Sys,       // OSから得た合計バイト数
	// 	// 	mem.HeapAlloc, // Allocと同値
	// 	// 	mem.HeapSys,
	// 	// 	mem.HeapReleased, // OSへ返却されたヒープ
	// 	// )
	// 	// time.Sleep(5 * time.Second)
	// 	// // fmt.Println("Executed gc")
	// 	// runtime.GC()
	// 	// debug.FreeOSMemory()
	// 	// if mem.Alloc > 3000000 {
	// 	// 	runtime.GC()
	// 	// 	debug.FreeOSMemory()
	// 	// }
	// }
}

func main() {

	// *previousLine = 0

	// // コマンドの実行結果をPipeで受け取る
	// command := exec.Command("php", "-i")
	// buffer, err := command.StdoutPipe()

	// if err != nil {
	// 	panic(err)
	// }
	// command.Start()

	// wiredrawing.LoadBuffer(buffer, previousLine)

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
		// var echo = fmt.Print
		var code int = 0
		for {
			code, _ = <-exit

			if code == 1 {
				os.Exit(code)
			} else if code == 4 {
				fmt.Print("[Ignored interrupt].\r\n")
			} else {
				if runtime.GOOS != "darwin" {
					fmt.Print("[Ignored interrupt].\r\n")
				}
			}
		}
	}(exit)

	// ----------------------------------------------
	// 標準入力を可能にする
	// 標準入力の開始
	// ----------------------------------------------
	// var waiter *sync.WaitGroup = new(sync.WaitGroup)
	// waiter.Add(1)
	// go inputter.StandByInput(waiter)
	// waiter.Wait()
	inputter.StandByInput()
}
