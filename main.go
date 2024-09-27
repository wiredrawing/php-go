// build option go build -ldflags "-w -s"  -trimpath
package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	//"golang.org/x/sys/windows"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"phpgo/cmd"
	"phpgo/config"
	"phpgo/errorhandler"
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
	_, err = h.Write(readBytes)
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
	fmt.Printf("currentDir: %v\r\n", currentDir)
	//// 過去に起動した一時ファイルを削除する
	//
	//globs, err := filepath.Glob(currentDir + "./PHP_GO_validation*")
	//if err != nil {
	//	panic(err)
	//}
	//for _, glob := range globs {
	//	_ = os.Remove(glob)
	//}

	// ユーザーが入力したファイル
	manualFp, err := os.OpenFile(filePathForSurveillance, os.O_CREATE|os.O_RDWR, 0777)
	size, err := manualFp.Write([]byte("<?php\n"))
	_ = manualFp.Close()
	// エラーの場合は以下の関数内で終了するため、エラー処理は不要
	errorhandler.ErrorHandler(err)
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
	var _ bool
	var err error
	// surveillanceモードの場合に常時開くエディタを指定する
	// デフォルトは .env ファイルを探索する
	err = godotenv.Load()
	if err != nil {
		fmt.Println(inputter.ColorWrapping(config.Red, "環境変数がロードできません。ターミナルのみ起動します。"))
	}
	err = godotenv.Load(".phpgo")
	if err != nil {
		fmt.Println(inputter.ColorWrapping(config.Red, "環境変数がロードできません。ターミナルのみ起動します。"))
	}
	var editorPath = os.Getenv("EDITOR_PATH")

	commandLineOption := cmd.Execute()

	if commandLineOption.Help || commandLineOption.Version || commandLineOption.Toggle {
		os.Exit(0)
	}
	// コマンドライン引数を取得
	var phpPath *string = stringp(commandLineOption.Phppath)
	var surveillanceFile *string = stringp(commandLineOption.Surveillance)
	var prompt *string = stringp(commandLineOption.Prompt)
	var saveFileName *string = stringp(commandLineOption.SaveFileName)
	var editorPathFromCommandLineOption *string = stringp(commandLineOption.EditorPath)

	// phpの実行パスを設定
	if *phpPath == "-" {
		// デフォルトのままの場合は<php>に設定
		*phpPath = "php"
	}

	// 割り込み監視用
	var signalChan = make(chan os.Signal)

	if *surveillanceFile != DefaultSurveillanceFileName && *surveillanceFile != "" {
		// 事前に過去に作成された一時ファイルを削除する
		tdir, err := os.MkdirTemp("", "phpgo")
		fmt.Printf("tdir: %v\r\n", tdir)
		if err != nil {
			panic(err)
		}
		targetPath := filepath.Join(tdir, *surveillanceFile+".php")
		tempF, err := os.Create(targetPath)
		if err != nil {
			panic(err)
		}
		var filePathForSurveillance string = tempF.Name()
		_ = tempF.Close()
		// Create new watcher.
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}

		defer func(watcher *fsnotify.Watcher) {
			err := watcher.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(watcher)

		// 指定されたエディタパスでファイルを開く
		// もしエディタパスが指定されていなければスルーする
		fmt.Printf("editorPathfromCommandLineOption: %v\r\n", *editorPathFromCommandLineOption)
		if len(editorPath) > 0 {
			err = exec.Command(editorPath, filePathForSurveillance).Start()
			if (err != nil) && (err.Error() != "exit status 1") {
				panic(err)
			}
		} else if len(*editorPathFromCommandLineOption) > 0 {
			err = exec.Command(*editorPathFromCommandLineOption, filePathForSurveillance).Start()
			if (err != nil) && (err.Error() != "exit status 1") {
				panic(err)
			}
		}
		// Start listening for events.
		go ExecuteSurveillanceFile(watcher, filePathForSurveillance, phpPath)
		if err := watcher.Add(tdir); err != nil {
			log.Fatal(err)
		}
	}

	// コンソールの監視
	signal.Notify(
		signalChan,
		os.Interrupt,
		os.Kill,
		//windows.SIGKILL,
		//windows.SIGHUP,
		//windows.SIGINT,
		//windows.SIGTERM,
		//windows.SIGQUIT,
		//windows.SIGTRAP,
		//windows.SIGSEGV,
		//windows.SIGABRT,
		//windows.Signal(0x0A),
		//windows.Signal(0x0B),
		//windows.Signal(0x0C),
		//windows.Signal(0x0D),
		//windows.Signal(0x0E),
		//windows.Signal(0x0F),
		//windows.Signal(0x10),
		//windows.Signal(0x13),
		//windows.Signal(0x14), // Windowsの場合 SIGTSTPを認識しないためリテラルで指定する
	)

	// GCを実行
	//go regularsGarbageCollection()

	var exit = make(chan int)
	// 割り込み対処を実行するGoルーチン
	go parallel.InterruptProcess(exit, signalChan)

	go func(exit chan int) {
		// var echo = fmt.Print
		var code = 0
		for {
			code = <-exit
			if code == 1 {
				os.Exit(code)
			} else if code == 4 {
				fmt.Printf(inputter.ColorWrapping(config.Yellow, "Please input the word 'exit' to exit the program.\r\n"))
				//fmt.Print("[Ignored interrupt].\r\n")
			} else {
				if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
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
				_, err = inputter.StandByInput(*phpPath, *prompt, *saveFileName)
			} else {
				fmt.Printf("%T\r\n", err)
				fmt.Println("errorの型アサーションに失敗")
			}
		}
	}()
	fmt.Println(inputter.ColorWrapping(config.Green, "[The applicaiton was just started.]"))
	_, err = inputter.StandByInput(*phpPath, *prompt, *saveFileName)
	if err != nil {
		panic(err)
	}

}
