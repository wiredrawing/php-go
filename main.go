package main

import (
	// 標準パッケージ

	sha2562 "crypto/sha256"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/sys/windows"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"php-go/wiredrawing"
	"php-go/wiredrawing/parallel"
	"runtime"

	// ここは独自パッケージ

	// _をつけた場合は パッケージ内のinit関数のみ実行される

	"php-go/wiredrawing/inputter"
	//"golang.org/x/sys/windows"
)

// var command *cobra.Command = new(cobra.Command)

// 割り込み監視用
var signal_chan chan os.Signal = make(chan os.Signal)

// ガベージコレクションを任意の時間間隔で実行
func regularsGarbageCollection() {

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

var _ bool
var err error
var targetFileName string = ""
var workingDirectory string = ""

// 監視対象のファイル名
var filePathForSurveillance string = ""

func main() {

	// []string型でコマンドライン引数を受け取る
	var arguments = os.Args
	// もしファイル名が指定されている場合はファイル監視処理に入る

	if arrayIndexExists(arguments, 1) {
		targetFileName = arguments[1]
		workingDirectory, err = os.Getwd()
		fmt.Println(workingDirectory)
		if err != nil {
			// 作業ディレクトリが取得できない場合
			panic(err)
		}
		filePathForSurveillance = workingDirectory + "\\" + targetFileName
		_, err = os.Stat(filePathForSurveillance)
		if err == nil {
			// ファイルが既に存在する場合は内容をtruncateする
			_ = os.Truncate(filePathForSurveillance, 0)
		} else {
			// ファイルが存在しない場合は新規作成する
			_, err := os.Create(filePathForSurveillance)
			if err != nil {
				panic(err)
			}
		}

		// Create new watcher.
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		// Start listening for events.
		go func() {
			var previousHash [32]byte
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Has(fsnotify.Write) && event.Name == filePathForSurveillance {
						surveillanceFile, _ := os.Open(event.Name)
						readByte, _ := ioutil.ReadAll(surveillanceFile)
						sha256 := sha2562.Sum256(readByte)
						if sha256 == previousHash {
							continue
						}
						// 古いハッシュを更新
						previousHash = sha256
						command := exec.Command("php", filePathForSurveillance)
						buffer, _ := command.StdoutPipe()
						err := command.Start()
						if err != nil {
							return
						}
						var previousLine *int = new(int)
						*previousLine = 0
						_, err2 := wiredrawing.LoadBuffer(buffer, previousLine, true, false)
						if err2 != nil {
							return
						}
						fmt.Fprint(os.Stdout, "\n")
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				}
			}
		}()
		watcher.Add(workingDirectory)
	}

	// *previousLine = 0

	// // コマンドの実行結果をPipeで受け取る
	// command := exec.Command("php", "-i")
	// buffer, err := command.StdoutPipe()

	// if err != nil {
	// 	panic(err)
	// }
	// command.Start()

	// wiredrawing.LoadBuffer(buffer, previousLine)

	//// ディレクトリのwatcherを作成
	//watcher, watcherErr := fsnotify.NewWatcher()
	//if watcherErr != nil {
	//	panic(watcherErr)
	//}
	//defer func(watcher *fsnotify.Watcher) {
	//	err := watcher.Close()
	//	if err != nil {
	//		panic(err)
	//	}
	//}(watcher)
	//
	//go func() {
	//	for {
	//		select {
	//		case event, ok := <-watcher.Events:
	//			if !ok {
	//				return
	//			}
	//			if event.Has(fsnotify.Write) {
	//				fmt.Println("書き込みイベント発生")
	//			}
	//		case err := <-watcher.Errors:
	//			fmt.Println(err)
	//		}
	//	}
	//
	//}()

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
	go regularsGarbageCollection()

	var exit chan int = make(chan int)
	// 割り込み対処を実行するGoルーチン
	go parallel.InterruptProcess(exit, signal_chan)

	go func(exit chan int) {
		// var echo = fmt.Print
		var code int = 0
		for {
			code = <-exit

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
	_, err = inputter.StandByInput()
	if err != nil {
		panic(err)
	}
}

func arrayIndexExists(array []string, index int) bool {
	return index >= 0 && index < len(array)
}
