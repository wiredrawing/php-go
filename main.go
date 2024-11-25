// build option go build -ldflags "-w -s"  -trimpath
package main

import (
	"crypto/sha256"
	"fmt"
	//"github.com/fsnotify/fsnotify"
	//"golang.org/x/sys/windows"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"phpgo/cmd"
	"phpgo/config"
	//"phpgo/errorhandler"
	//"phpgo/wiredrawing"
	"phpgo/wiredrawing/inputter"
	"phpgo/wiredrawing/parallel"
	"runtime"
	// ここは独自パッケージ
	//"golang.org/x/sys/windows"
)

// var command *cobra.Command = new(cobra.Command)

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

func stringp(s string) *string {
	return &s
}

func main() {
	var _ bool
	var err error

	commandLineOption := cmd.Execute()

	if commandLineOption.Help || commandLineOption.Version || commandLineOption.Toggle {
		os.Exit(0)
	}
	// コマンドライン引数を取得
	var phpPath *string = stringp(commandLineOption.Phppath)
	_ = stringp(commandLineOption.Surveillance)
	var prompt *string = stringp(commandLineOption.Prompt)
	var saveFileName *string = stringp(commandLineOption.SaveFileName)
	_ = stringp(commandLineOption.EditorPath)

	// phpの実行パスを設定
	if *phpPath == "-" {
		// デフォルトのままの場合は<php>に設定
		*phpPath = "php"
	}

	// 割り込み監視用
	var signalChan = make(chan os.Signal)

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
				fmt.Printf(config.ColorWrapping(config.Yellow, "Please input the word 'exit' to exit the program.\r\n"))
				//fmt.Print("[Ignored interrupt].\r\n")
			} else {
				if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
					fmt.Print("[Ignored interrupt].\r\n")
				}
			}
			fmt.Printf("code: %v\r\n", code)
		}
	}(exit)
	// ----------------------------------了する場合はもう一度 ctrl+C を押して下さ------------
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
	//fmt.Println(inputter.ColorWrapping(config.Green, "[The applicaiton was just started.]"))
	_, err = inputter.StandByInput(*phpPath, *prompt, *saveFileName)
	if err != nil {
		panic(err)
	}

}
