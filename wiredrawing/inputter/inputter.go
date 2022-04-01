package inputter

import (
	"bufio"
	"fmt"
	"go-sample/wiredrawing"
	"os"
	"os/exec"
	"sync"
)

// 入力用ポインタ
var scanner *bufio.Scanner

var logFile string = ".log.dat"

// 改行文字を定義
const newLine string = "\n"

var command *exec.Cmd
var previousLine *int = new(int)

// 標準入力を待ち受ける関数
func StandByInput(waiter *sync.WaitGroup) {

	*previousLine = 0
	// 標準入力の内容を保存する用のファイルポインタを作成
	var file *os.File
	var err error
	file, err = os.Create(logFile)
	if err != nil {
		panic(err)
	}

	// phpの<?phpタグを記述する
	file.WriteString("<?php " + "\n")

	// 遅延実行
	defer file.Close()

	// ----------------------------------------------
	// 標準入力を可能にする
	// 標準入力の開始
	// ----------------------------------------------
	scanner = bufio.NewScanner(os.Stdin)

	var inputText string = ""
	for {
		fmt.Print("  >> ")
		var isOk bool = scanner.Scan()
		if isOk != true {
			fmt.Println("scanner.Scan()が失敗")
			// scannerを初期化
			scanner = nil
			scanner = bufio.NewScanner(os.Stdin)
			continue
		}
		inputText = scanner.Text()
		// 入力内容が exit ならアプリケーションを終了
		if len(inputText) > 0 {
			if inputText == "exit" {
				waiter.Done()
				// os.Exit(1)
			} else if inputText == "clear" || inputText == "refresh" {
				// 入力内容が [clear] or [refresh] だった場合は入力内容をクリア
				// ファイルサイズを空にする
				err = os.Truncate(logFile, 0)
				file.Seek(0, 0)
			} else {
				// プログラムの実行処理
				// ファイルポインタへ書き込み
				file.WriteString(inputText + newLine)
				// phpでの改行を追加する
				file.WriteString("echo(PHP_EOL);")

				// 一旦入力内容が正しく終了するかどうかを検証
				command = exec.Command("php", logFile)
				command.Run()
				exitCode := command.ProcessState.ExitCode()
				// 終了コードが0でない場合は何かしらのエラー
				if exitCode != 0 {
					panic("phpコマンドの実行に失敗")
				}

				command = exec.Command("php", logFile)
				// fmt.Println(command)
				buffer, err := command.StdoutPipe()

				if err != nil {
					panic(err)
				}
				// bufferの読み取り開始
				command.Start()
				wiredrawing.LoadBuffer(buffer, previousLine)
			}
		}
	}
	// 標準入力の終了
}
