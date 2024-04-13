// build option go build -ldflags "-w -s"  -trimpath
package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/sys/windows"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"php-go/wiredrawing"
	"php-go/wiredrawing/parallel"
	"runtime"
	"time"

	// ここは独自パッケージ

	// _をつけた場合は パッケージ内のinit関数のみ実行される

	"php-go/wiredrawing/inputter"
	//"golang.org/x/sys/windows"
)

// var command *cobra.Command = new(cobra.Command)

// ガベージコレクションを任意の時間間隔で実行
func regularsGarbageCollection() {

	var mem runtime.MemStats
	for {
		runtime.ReadMemStats(&mem)
		//fmt.Printf("(1)Alloc:%d, (2)TotalAlloc:%d, (3)Sys:%d, (4)HeapAlloc:%d, (5)HeapSys:%d, (6)HeapReleased:%d\r\n",
		//	mem.Alloc, // HeapAllocと同値
		//	mem.TotalAlloc,
		//	mem.Sys,       // OSから得た合計バイト数
		//	mem.HeapAlloc, // Allocと同値
		//	mem.HeapSys,
		//	mem.HeapReleased, // OSへ返却されたヒープ
		//)
		time.Sleep(5 * time.Second)
		// fmt.Println("Executed gc")
		runtime.GC()
		//debug.FreeOSMemory()
		if mem.Alloc > 1000000 {
			runtime.GC()
			//debug.FreeOSMemory()
		}
	}
}

// ファイルのハッシュ値を計算する
func hash(filepath string) string {
	file, err := os.Open(filepath)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	h := sha256.New()
	readBytes, _ := ioutil.ReadAll(file)
	h.Write(readBytes)
	hashedValue := h.Sum(nil)
	return string(hashedValue)
}

const DefaultSurveillanceFileName string = "-"

func ExecuteSurveillanceFile(watcher *fsnotify.Watcher, filePathForSurveillance string, phpPath *string) {
	// ファイル内容のハッシュ計算用に保持
	//var previousHash []byte
	//var hashedValue []byte
	var hashedValue string
	var previousHash string
	// forループ外で宣言する
	var php = wiredrawing.PhpExecuter{
		PhpPath: *phpPath,
	}
	// 一時ファイルの作成
	currentDir, _ := os.Getwd()

	// 過去に起動した一時ファイルを削除する

	globs, err := filepath.Glob(currentDir + "./PHP_GO_validation*")
	if err != nil {
		panic(err)
	}
	for _, glob := range globs {
		_ = os.Remove(glob)
	}
	// 起動時点の一時ファイルを作成
	fpForValidate, err := os.CreateTemp(currentDir, "PHP_GO_validation.php")
	defer func() {
		err := fpForValidate.Close()
		if err != nil {
			panic(err)
		}
	}()
	if err != nil {
		panic(err)
	}
	//fmt.Printf("fpForValidate.Name(): %s\r\n", fpForValidate.Name())

	// ユーザーが入力したファイル
	manualFp, err := os.Open(filePathForSurveillance)
	if err != nil {
		panic(err)
	}
	var previousExcecuteCode []string
	for {
		// php実行時の出力行数を保持
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) && event.Name == filePathForSurveillance {
				hashedValue = hash(filePathForSurveillance)
				//surveillanceFile, _ := os.Open(event.Name)
				//readByte, _ := ioutil.ReadAll(surveillanceFile)
				//h := sha256.New()
				//h.Write(readByte)
				//hashedValue = h.Sum(nil)
				//log.Println("Hashed value: ", hashedValue)
				if hashedValue == previousHash {
					continue
				}
				// 古いハッシュを更新
				previousHash = hashedValue
				if len(previousExcecuteCode) > 0 {
					nextExecutedCode, _ := wiredrawing.File(filePathForSurveillance)
					if len(previousExcecuteCode) >= len(nextExecutedCode) {
						for index := 0; index < len(previousExcecuteCode); index++ {
							if previousExcecuteCode[index] != nextExecutedCode[index] {
								php.SetPreviousList(0)
								break
							}
						}
					}
					previousExcecuteCode = nextExecutedCode
				} else {
					previousExcecuteCode, _ = wiredrawing.File(filePathForSurveillance)
				}
				io.Copy(fpForValidate, manualFp)
				php.SetOkFile(filePathForSurveillance)
				php.SetNgFile(filePathForSurveillance)
				// 致命的なエラー
				if bytes, err := php.DetectFatalError(); err != nil || len(bytes) > 0 {
					// Fatal Errorが検出された場合はエラーメッセージを表示して終了
					fmt.Println(inputter.ColorWrapping("31", string(bytes)))
					continue
				}
				if bytes, err := php.DetectErrorExceptFatalError(); (err != nil) || len(bytes) > 0 {
					fmt.Println(inputter.ColorWrapping("31", string(bytes)))
					continue
				}
				fmt.Println(" --> ")
				size, err := php.Execute()
				if err != nil {
					panic(err)
				}
				if size > 0 {
					fmt.Println("")
				}
				//command := exec.Command(*phpPath, filePathForSurveillance)
				//buffer, _ := command.StdoutPipe()
				//err := command.Start()
				//if err != nil {
				//	return
				//}
				//_, _ = wiredrawing.LoadBuffer(buffer, previousLine, true, false, "34")
				//fmt.Println("")
				//fmt.Printf("*previousLine: %d\r\n", *previousLine)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}
func main() {
	//go spinner()
	// コマンドライン引数を取得
	phpPath := flag.String("phppath", "-", "PHPの実行ファイルのパスを入力")
	surveillanceFile := flag.String("surveillance", DefaultSurveillanceFileName, "監視対象のファイル名を入力")
	flag.Parse()

	// phpの実行パスを設定
	if *phpPath == "-" {
		// デフォルトのままの場合は<php>に設定
		*phpPath = "php"
	}

	var _ bool
	var err error

	//
	// 監視対象のファイル名
	var filePathForSurveillance = ""

	// 割り込み監視用
	var signal_chan = make(chan os.Signal)

	// []string型でコマンドライン引数を受け取る
	var targetFileName = ""
	var workingDirectory = ""
	if *surveillanceFile != DefaultSurveillanceFileName && *surveillanceFile != "" {
		targetFileName = *surveillanceFile
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

		defer func(watcher *fsnotify.Watcher) {
			err := watcher.Close()
			if err != nil {
				panic(err)
			}
		}(watcher)

		// Start listening for events.
		go ExecuteSurveillanceFile(watcher, filePathForSurveillance, phpPath)
		//go func() {
		//	// ファイル内容のハッシュ計算用に保持
		//	//var previousHash []byte
		//	//var hashedValue []byte
		//	var hashedValue string
		//	var previousHash string
		//	for {
		//		select {
		//		case event, ok := <-watcher.Events:
		//			if !ok {
		//				return
		//			}
		//			if event.Has(fsnotify.Write) && event.Name == filePathForSurveillance {
		//				hashedValue = hash(filePathForSurveillance)
		//				//surveillanceFile, _ := os.Open(event.Name)
		//				//readByte, _ := ioutil.ReadAll(surveillanceFile)
		//				//h := sha256.New()
		//				//h.Write(readByte)
		//				//hashedValue = h.Sum(nil)
		//				log.Println("Hashed value: ", hashedValue)
		//				if hashedValue == previousHash {
		//					continue
		//				}
		//				// 古いハッシュを更新
		//				previousHash = hashedValue
		//				//
		//				command := exec.Command(*phpPath, filePathForSurveillance)
		//				buffer, _ := command.StdoutPipe()
		//				err := command.Start()
		//				if err != nil {
		//					return
		//				}
		//				var previousLine *int = new(int)
		//				_, _ = wiredrawing.LoadBuffer(buffer, previousLine, true, false, "34")
		//				fmt.Println("")
		//				fmt.Printf("*previousLine: %d\r\n", *previousLine)
		//			}
		//		case err, ok := <-watcher.Errors:
		//			if !ok {
		//				return
		//			}
		//			log.Println("error:", err)
		//		}
		//	}
		//}()
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
	//go regularsGarbageCollection()

	var exit = make(chan int)
	// 割り込み対処を実行するGoルーチン
	go parallel.InterruptProcess(exit, signal_chan)

	go func(exit chan int) {
		// var echo = fmt.Print
		var code = 0
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
	defer func() {
		if err := recover(); err != nil {
			if err, ok := err.(error); ok {
				fmt.Println(err)
				_, err = inputter.StandByInput(*phpPath)
			} else {
				fmt.Println("errorの型アサーションに失敗")
			}

		}
	}()
	_, err = inputter.StandByInput(*phpPath)
	if err != nil {
		panic(err)
	}

}

func arrayIndexExists(array []string, index int) bool {
	return index >= 0 && index < len(array)
}
