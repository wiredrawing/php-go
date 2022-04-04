package inputter

import (
	"bufio"
	"fmt"
	"go-sample/wiredrawing"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

// 入力用ポインタ
var scanner *bufio.Scanner

var ngFile string = ".validation.dat"

var okFile string = ".success.dat"

// 改行文字を定義
const newLine string = "\n"

var command *exec.Cmd
var previousLine *int = new(int)

// 入力内容を保持する変数
var inputText string = ""

var file1 *os.File
var file2 *os.File

var err error

// ----------------------------------------------//
// パッケージの初期化
// init関数は値を返却できない
// ----------------------------------------------
func init() {
	// 入力内容のコマンド結果確認用
	file1, err = os.Create(ngFile)
	if err != nil {
		panic(err)
	}
	// phpの<?phpタグを記述する
	file1.WriteString("<?php " + "\n")

	// 実際の実行ファイル用
	file2, err = os.Create(okFile)
	if err != nil {
		panic(err)
	}
	// phpの<?phpタグを記述する
	file2.WriteString("<?php " + "\n")
}

// 標準入力を待ち受ける関数
func StandByInput() {

	*previousLine = 0

	// 遅延実行
	defer file1.Close()
	defer file2.Close()

	// ----------------------------------------------
	// 標準入力を可能にする
	// 標準入力の開始
	// ----------------------------------------------
	scanner = bufio.NewScanner(os.Stdin)
	var prompt string = " >> "
	for {
		file1, err = os.OpenFile(ngFile, os.O_APPEND|os.O_WRONLY, 0777)
		file2, err = os.OpenFile(okFile, os.O_APPEND|os.O_WRONLY, 0777)
		fmt.Print(prompt)
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
		if len(inputText) == 0 {
			continue
		}

		// ---------------------------------------------
		// exitコマンドの場合はシェルを終了
		// ---------------------------------------------
		if inputText == "exit" {
			os.Exit(1)
			break
		}

		// ---------------------------------------------
		// スペースで分割して delete indexNumber を取り出す
		// ---------------------------------------------
		tokens := strings.Split(inputText, " ")
		if tokens[0] == "delete" {
			{
				file1.Close()
				file2.Close()
				var fileBuffer []string
				file, err := os.Open(ngFile)
				if err != nil {
					panic(err)
				}
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					{
						fileBuffer = append(fileBuffer, scanner.Text())
					}
				}
				var indexToDelete int
				indexToDelete, err = strconv.Atoi(tokens[1])
				if err != nil {
					panic(err)
				}

				fileBuffer = append(fileBuffer[:indexToDelete], fileBuffer[indexToDelete+1:]...)
				file.Close()

				// validation用ファイルを空に
				file1.Close()
				os.Truncate(ngFile, 0)
				file, err = os.OpenFile(ngFile, os.O_APPEND|os.O_WRONLY, 0777)
				if err != nil {
					panic(err)
				}
				file.Seek(0, 0)
				for _, value := range fileBuffer {
					// 行末に改行文字を入力
					_, err := file.WriteString(value + "\n")
					if err != nil {
						panic(err)
					}
				}
				file.Close()

				// 実行用用ファイルを空に
				// ngFile => okFile に内容をコピ
				file2.Close()
				os.Truncate(okFile, 0)
				file, err = os.OpenFile(okFile, os.O_APPEND|os.O_WRONLY, 0777)
				if err != nil {
					panic(err)
				}
				file.Seek(0, 0)
				for _, value := range fileBuffer {
					// 行末に改行文字を入力
					_, err := file.WriteString(value + "\n")
					if err != nil {
						panic(err)
					}
				}
				file.Close()

				continue
			}

		}

		// 入力内容が [clear] or [refresh] だった場合は入力内容をクリア
		// ファイルサイズを空にする
		if inputText == "clear" || inputText == "refresh" {
			// phpスクリプトチェック用ファイルを殻にする
			os.Truncate(ngFile, 0)
			file1.Seek(0, 0)
			f, err := os.Open(okFile)
			if err != nil {
				panic(err)
			}
			source, err := ioutil.ReadAll(f)
			if err != nil {
				panic(err)
			}
			file1.WriteString(string(source))
			prompt = " >> "
			runtime.GC()
			debug.FreeOSMemory()
			continue
		}

		// cat と入力すると現在まで入力している内容を出力する
		if inputText == "cat" {

			(func() {
				catFile, err := os.Open(ngFile)
				if err != nil {
					panic(err)
				}
				tempScanner := bufio.NewScanner(catFile)

				var index int = 0
				var indexStr string = ""
				for tempScanner.Scan() {
					indexStr = fmt.Sprintf("%03d", index)
					fmt.Print(indexStr + ": ")
					fmt.Println(tempScanner.Text())
					index++
				}
			})()

			continue
		}

		// プログラムの実行処理
		// ファイルポインタへ書き込み
		file1.WriteString(inputText + newLine)
		// phpでの改行を追加する
		// file.WriteString("echo(PHP_EOL);")

		// 一旦入力内容が正しく終了するかどうかを検証
		command = exec.Command("php", ngFile)
		command.Run()
		exitCode := command.ProcessState.ExitCode()

		// --------------------------------------------
		// 終了コードが0でない場合は何かしらのエラー
		// あるいは入力途中とする
		// --------------------------------------------
		if exitCode != 0 {
			prompt = " ... "
			continue
		}
		prompt = " >> "

		// 正常終了の場合は ngFile中身をokFileにコピー
		f, err := os.Open(ngFile)
		if err != nil {
			panic(err)
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}
		os.Truncate(okFile, 0)
		file2.Seek(0, 0)
		file2.WriteString(string(b))
		command = exec.Command("php", okFile)
		// fmt.Println(command)
		buffer, err := command.StdoutPipe()

		if err != nil {
			panic(err)
		}
		// bufferの読み取り開始
		command.Start()
		wiredrawing.LoadBuffer(buffer, previousLine)
		fmt.Println("")
	}
	// 標準入力の終了
}
