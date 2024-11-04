package inputter

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"phpgo/config"
	"phpgo/wiredrawing"
	"runtime"
	"runtime/debug"
	"strings"
)

// 入力内容を保持する変数
//var inputText = ""

var err error

func makeDirectory(dir string) bool {
	_, err := os.Stat(dir)
	// 指定したディレクトリが存在している場合は何もしない
	if (err != nil) && os.IsNotExist(err) {
		// <dotDir>が存在しない場
		err = os.Mkdir(dir, 0777)
		if err != nil {
			panic(err)
		}
	}
	return true
}

func concatenate(fileName string) []string {
	catFile, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	tempScanner := bufio.NewScanner(catFile)

	var index = 0
	var indexStr = ""
	fmt.Println("")
	var outputList []string = make([]string, 0, 100)
	var t string = ""
	for tempScanner.Scan() {
		indexStr = fmt.Sprintf("%03d", index)
		fmt.Print(ColorWrapping("34", indexStr) + ": ")
		t = tempScanner.Text()
		outputList = append(outputList, t)
		fmt.Println(ColorWrapping("32", t))
		index++
	}
	fmt.Println("")
	return outputList
}

// ----------------------------------------------//
// パッケージの初期化
// init関数は値を返却できない
// ----------------------------------------------
func init() {

	const ngFile = ".validation.dat"
	const okFile = ".success.dat"
	var filePathForError = ".erorr_message.dat"

	var homeDir string
	homeDir, _ = os.UserHomeDir()
	// 本アプリケーション専用の設定ディレクトリ
	var dotDir string = ""
	if runtime.GOOS == "windows" {
		dotDir = homeDir + "\\.php-go"
	} else {
		dotDir = homeDir + "/.php-go"
	}

	// ディレクトリが存在しない場合は作成する
	makeDirectory(dotDir)

	filePathForError = dotDir + "\\" + filePathForError

	var file1 *os.File
	var file2 *os.File
	var file3 *os.File

	// 入力内容のコマンド結果確認用
	file1, err = os.Create(dotDir + "\\" + ngFile)
	if err != nil {
		panic(err)
	}
	// phpの<?phpタグを記述する
	file1.Write([]byte("<?php " + "\n"))

	// 実際の実行ファイル用
	file2, err = os.Create(dotDir + "\\" + okFile)
	if err != nil {
		panic(err)
	}
	// phpの<?phpタグを記述する
	file2.WriteString("<?php " + "\n")

	// 実行時エラーの出力用ファイル
	file3, err = os.Create(filePathForError)
	if err != nil {
		log.Printf("Pointer of file3: %p\n", file3)
		log.Printf("Could not create the file: %s", filePathForError)
		panic(err)
	}

}

// StandByInput 標準入力を待ち受ける関数
func StandByInput(phpPath string, inputPrompt string, saveFileName string) (bool, error) {
	// phpPathが未指定の場合はハイフンが設定されているため
	// それ以外の場合は指定したphpパスで実行する
	var phpExecutePath = "php"
	if phpPath != "-" {
		phpExecutePath = phpPath
	}
	// ターミナルを終了させるためのキーワード群
	//var wordsToExit = make([]string, 0, 32)
	var wordsToExit []string = []string{
		"y",
		"Y",
		"yes",
	}

	var ngFile = ".validation.dat"
	var okFile = ".success.dat"

	//dname, _ := os.MkdirTemp("", "phpgo_")
	////fmt.Printf("tempDir: %v\r\n", dname)
	//ngTempF, err := os.CreateTemp(dname, "ngFile.dat")
	//if err != nil {
	//	panic(err)
	//}
	//okTempF, err := os.CreateTemp(dname, "okFile.dat")
	//if err != nil {
	//	panic(err)
	//}
	//
	//ngFile = ngTempF.Name()
	//ngTempF.Close()
	//okFile = okTempF.Name()
	//okTempF.Close()
	//filePathForError = dotDir + "\\" + filePathForError

	// OSの一時ファイル作成に任せる
	okFileTemp, err := os.CreateTemp("", ".success.dat.")
	if err != nil {
		panic(err)
	}
	// <?php を 一行目に書き込む
	_, _ = okFileTemp.Write([]byte("<?php " + "\n"))
	okFile = okFileTemp.Name()

	ngFileTemp, err := os.CreateTemp("", ".validation.dat.")
	if err != nil {
		panic(err)
	}
	// <?php を 一行目に書き込む
	_, _ = ngFileTemp.Write([]byte("<?php " + "\n"))
	ngFile = ngFileTemp.Name()

	// ----------------------------------------------
	// 標準入力を可能にする
	// 標準入力の開始
	// ----------------------------------------------
	var prompt = fmt.Sprintf(" %s ", inputPrompt)
	var inputText string
	//var previousInputText = ""

	// forループ外で宣言する
	var php = wiredrawing.PhpExecuter{
		PhpPath: phpExecutePath,
	}
	php.InitDB()

	// ヘルプメッセージを定義
	var helpMessages = make([]string, 0, 32)
	helpMessages = append(helpMessages, fmt.Sprintf("[help]"))
	helpMessages = append(helpMessages, fmt.Sprintf("clear:    入力途中の内容を破棄します."))
	helpMessages = append(helpMessages, fmt.Sprintf("rollback: 入力済みの入力を一行ずつ削除します."))
	helpMessages = append(helpMessages, fmt.Sprintf("cat:      入力済みの内容を表示します."))
	helpMessages = append(helpMessages, fmt.Sprintf("exit:     アプリケーションを終了します."))
	helpMessages = append(helpMessages, fmt.Sprintf("errors:   過去のエラーを全て表示します."))
	var helpMessage = strings.Join(helpMessages, "\n")

	var rawInputText string = ""
	for {
		php.SetOkFile(okFile)
		php.SetNgFile(ngFile)
		if php.IsPermissibleError == true {
			fmt.Print("\033[33m")
			fmt.Print(ColorWrapping("33", prompt))
		} else {
			//fmt.Print(ColorWrapping("0", prompt))
		}
		// 両端のスペースを削除
		rawInputText = wiredrawing.StdInput("", rawInputText)
		//previousInputText = inputText
		inputText = strings.TrimSpace(rawInputText)

		if len(inputText) == 0 {
			runtime.GC()
			continue
		}

		if inputText == "help" {
			fmt.Println(ColorWrapping("33", helpMessage))
			continue
		}

		if inputText == "errors" {
			var wholeErrors = php.WholeErrors()
			for key, value := range wholeErrors {
				fmt.Print(ColorWrapping(config.Green, fmt.Sprintf("[%03d] => ", key+1)))
				fmt.Print(ColorWrapping(config.Red, value))
			}
			continue
		}
		if inputText == "reset errors" {
			php.ResetWholeErrors()
			continue
		}
		if inputText == "rollback" {
			//lines, err := wiredrawing.File(ngFile)
			//if err != nil {
			//	panic(err)
			//}
			//if len(lines) == 1 {
			//	continue
			//}
			//_ = popStirngToFile(ngFile, -1)
			php.Rollback()
			isFatal, _ := php.DetectFatalError()
			errorBuffer := php.ErrorBuffer
			if isFatal == true {
				prompt = " ... "
			} else {
				if len(errorBuffer) > 0 {
					prompt = " ... "
				} else {
					prompt = " " + (inputPrompt) + " "
				}
			}
			//php.SetOkFile(ngFile)
			cl := php.SetPreviousList(0)
			_, _ = php.Execute(false)
			//_ = concatenate(ngFile)
			logs := php.Cat()
			for index := range logs {
				indexStr := fmt.Sprintf("%04d", logs[index]["id"])
				fmt.Print(ColorWrapping("34", indexStr) + ": ")
				fmt.Println(ColorWrapping("32", (logs[index]["text"]).(string)))
			}
			php.SetPreviousList(cl)
			continue
		}

		if inputText == "save" {
			php.Save(saveFileName)
			continue
		}
		if inputText == "exit" {
			// コンソールを終了するための標準入力を取得する

			fmt.Println(ColorWrapping("31", "[PHPコマンドラインを終了します。本当に終了する場合は<yes>と入力して下さい。]"))
			//fmt.Print(prompt)
			// 両端のスペースを削除
			var inputText = wiredrawing.StdInput(prompt, rawInputText)
			inputText = strings.TrimSpace(inputText)
			// 入力内容が空文字の場合コマンドラインを終了する
			if len(inputText) == 0 {
				fmt.Println(ColorWrapping("31", "キャンセルしました。"))
				continue
			}

			if wiredrawing.InArray(inputText, wordsToExit) {
				// 終了メッセージを表示
				// string型を[]byteに変換して書き込み
				var messageToEnd = []byte("Thank you for using me! Good by.")
				_, err := os.Stdout.Write([]byte(ColorWrapping("34", string(messageToEnd))))
				os.Stdout.Write([]byte("\n\n"))
				if err != nil {
					return false, err
				}
				break
			}
			continue
		}

		// 入力内容が [clear] or [refresh] だった場合は入力内容をクリア
		// ファイルサイズを空にする
		if inputText == "clear" {
			// バリデーション用のファイルを空にする
			php.Clear()
			// ただPHP開始タグのみ先頭に追記しておく
			php.WriteToNg("<?php " + "\n")
			fmt.Print("\033[0m")
			php.IsPermissibleError = false
			prompt = fmt.Sprintf(" %s ", inputPrompt)
			runtime.GC()
			debug.FreeOSMemory()
			continue
		}

		// cat と入力すると現在まで入力している内容を出力する
		if inputText == "cat" {
			logs := php.Cat()
			for index := range logs {
				indexStr := fmt.Sprintf("%04d", logs[index]["id"])
				fmt.Print(ColorWrapping("34", indexStr) + ": ")
				fmt.Println(ColorWrapping("32", (logs[index]["text"]).(string)))
			}
			//concatenate(ngFile)
			continue
		}

		php.WriteToNg(rawInputText + "\n")
		// ----------------------------------------------------------------
		// Fatal Error以外を検出する
		// ----------------------------------------------------------------
		isFatal, err := php.DetectFatalError()
		if err != nil {
			panic(err)
		}

		// FatalErrorの場合
		if isFatal == true {
			// PHP オブジェクトからErrorBufferを取り出す
			errorBuffer := php.ErrorBuffer
			if len(errorBuffer) > 0 {
				if php.IsPermissibleError == true {
					prompt = " ... "
				} else {
					// Fatal Errorが検出された場合はエラーメッセージを表示して終了
					fmt.Println(ColorWrapping("31", string(errorBuffer)))
					// 事前検証用のfileの中身を本実行用fileの中身と同じにする
					php.Rollback()
				}
				continue
			}
		} else if (isFatal == false) && len(php.ErrorBuffer) > 0 {
			// Fatal Error ではないがErrorBufferが空ではない場合
			fmt.Println(ColorWrapping("33", string(php.ErrorBuffer)))
			// 事前検証用のfileの中身を本実行用fileの中身と同じにする
			php.Rollback()
			continue
		}

		//_, _ = php.CopyFromNgToOk()

		if outputSize, err := php.Execute(true); err != nil {
			fmt.Println(ColorWrapping("31", err.Error()))
		} else {
			if outputSize > 0 {
				fmt.Print(ColorWrapping("0", "\n"))
			} else {
				fmt.Print(ColorWrapping("0", ""))
			}
		}
		rawInputText = ""
		prompt = fmt.Sprintf(" %s ", inputPrompt)
		continue
	}
	// 標準入力の終了

	return true, nil
}

func ColorWrapping(colorCode string, text string) string {
	return "\033[" + colorCode + "m" + text + "\033[0m"
}
