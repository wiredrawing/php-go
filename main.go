// build option go build -ldflags "-w -s"  -trimpath
package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/sys/windows"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"phpgo/cmd"
	"phpgo/wiredrawing"
	"phpgo/wiredrawing/inputter"
	"phpgo/wiredrawing/parallel"
	"runtime"
	"strings"
	"time"
	// ここは独自パッケージ
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
		//log.Println(err)
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
	fmt.Printf("readBytes: %v\r\n", readBytes)
	size, err := h.Write(readBytes)
	fmt.Printf("size: %v\r\n", size)
	if err != nil {
		log.Println(err)
		return ""
	}
	hashedValue := h.Sum(nil)
	var hashedValueString string = string(hashedValue)
	return hashedValueString
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

	// ユーザーが入力したファイル
	manualFp, err := os.OpenFile(filePathForSurveillance, os.O_CREATE|os.O_RDWR, 0777)
	size, err := manualFp.Write([]byte("<?php\n"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("size: %v\r\n", size)
	var previousExcecuteCode []string
	for {
		// php実行時の出力行数を保持
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) && event.Name != filePathForSurveillance {
				continue
			}
			hashedValue = hash(filePathForSurveillance)

			if hashedValue == previousHash {
				continue
			}
			// 古いハッシュを更新
			previousHash = hashedValue
			if len(previousExcecuteCode) > 0 {
				nextExecutedCode, _ := wiredrawing.File(filePathForSurveillance)

				for i := 0; i < len(previousExcecuteCode); i++ {
					if (len(nextExecutedCode) - 1) < i {
						php.SetPreviousList(0)
						break
					}
					// スペースは削除して比較
					p := strings.TrimSpace(previousExcecuteCode[i])
					n := strings.TrimSpace(nextExecutedCode[i])
					if p != n {
						php.SetPreviousList(0)
						break
					}
				}
				previousExcecuteCode = nextExecutedCode
			} else {
				previousExcecuteCode, _ = wiredrawing.File(filePathForSurveillance)
			}
			//io.Copy(fpForValidate, manualFp)
			php.SetOkFile(filePathForSurveillance)
			php.SetNgFile(filePathForSurveillance)
			// 致命的なエラー
			isFatal, err := php.DetectFatalError()
			if isFatal == true {
				// 致命的なエラーの場合
				// Fatal Errorが検出された場合はエラーメッセージを表示して終了
				fmt.Println("[Fatal Error]")
				fmt.Println(inputter.ColorWrapping("31", string(php.ErrorBuffer)))
				continue
			} else {
				if len(php.ErrorBuffer) > 0 {
					// 非Fatal Error
					fmt.Println("[None Fatal Error]")
					fmt.Println(inputter.ColorWrapping("33", string(php.ErrorBuffer)))
					continue
				}
			}
			if bytes, err := php.DetectErrorExceptFatalError(); (err != nil) || len(bytes) > 0 {
				fmt.Println(inputter.ColorWrapping("31", string(bytes)))
				continue
			}
			fmt.Println(" >>> ")
			size, err := php.Execute(true)
			if err != nil {
				panic(err)
			}
			if size > 0 {
				fmt.Println("")
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func stringp(s string) *string {
	return &s
}
func main() {
	commandConfig := cmd.Execute()

	var _ bool
	var err error

	//go spinner()
	// コマンドライン引数を取得
	var phpPath *string = stringp(commandConfig["phppath"])
	var surveillanceFile *string = stringp(commandConfig["surveillance"])
	var prompt *string = stringp(commandConfig["prompt"])
	var saveFileName *string = stringp(commandConfig["saveFileName"])
	//phpPath := flag.String("phppath", "-", "PHPの実行ファイルのパスを入力")
	//surveillanceFile := flag.String("surveillance", DefaultSurveillanceFileName, "監視対象のファイル名を入力")
	//flag.Parse()

	// phpの実行パスを設定
	if *phpPath == "-" {
		// デフォルトのままの場合は<php>に設定
		*phpPath = "php"
	}

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
		watcher.Add(workingDirectory)
	}

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
		windows.SIGTRAP,
		windows.SIGSEGV,
		windows.SIGABRT,
		windows.Signal(0x0A),
		windows.Signal(0x0B),
		windows.Signal(0x0C),
		windows.Signal(0x0D),
		windows.Signal(0x0E),
		windows.Signal(0x0F),
		windows.Signal(0x10),
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
				fmt.Printf(inputter.ColorWrapping("33", "Please input the word 'exit' to exit the program.\r\n"))
				//fmt.Print("[Ignored interrupt].\r\n")
			} else {
				if runtime.GOOS != "darwin" {
					fmt.Print("[Ignored interrupt].\r\n")
				}
			}
			fmt.Printf("code: %v\r\n", code)
		}
	}(exit)
	// ----------------------------------------------
	// 標準入力を可能にする
	// 標準入力の開始
	// ----------------------------------------------
	defer func() {
		if err := recover(); err != nil {
			if err, ok := err.(error); ok {
				fmt.Println(err)
				_, err = inputter.StandByInput(*phpPath, *prompt, *saveFileName, exit)
			} else {
				fmt.Printf("%T\r\n", err)
				fmt.Println("errorの型アサーションに失敗")
			}
		}
	}()
	fmt.Println(inputter.ColorWrapping("32", "[The applicaiton was just started.]"))
	_, err = inputter.StandByInput(*phpPath, *prompt, *saveFileName, exit)
	if err != nil {
		panic(err)
	}

}
