package inputter

import (
	"fmt"
	"os"
	"phpgo/config"
	"phpgo/wiredrawing"
	"runtime"

	//"runtime"
	"runtime/debug"
	"strings"
)

// ----------------------------------------------//
// パッケージの初期化
// init関数は値を返却できない
// ----------------------------------------------

// StandByInput 標準入力を待ち受ける関数
func StandByInput(phpPath string, inputPrompt string, saveFileName string) (bool, error) {
	//processId := os.Getpid()
	//fmt.Println("------------------------------------------------")
	//fmt.Printf("Process Id: %v \n", processId)
	//fmt.Println("入力したコードの実行: Alt + Enter or Input `_` ")
	//fmt.Println("直前までの入力内容を確認: Input `cat` ")
	//fmt.Println("直前の入力を取り消す: Input `rollback` ")
	//fmt.Println("シェルを終了する: Input `exit` then `yes` ")
	//fmt.Println("------------------------------------------------")

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

	//ngFileTemp, err := os.CreateTemp("", ".validation.dat.")
	//if err != nil {
	//	panic(err)
	//}
	//// <?php を 一行目に書き込む
	//_, _ = ngFileTemp.Write([]byte("<?php " + "\n"))
	//ngFile := ngFileTemp.Name()

	// ----------------------------------------------
	// 標準入力を可能にする
	// 標準入力の開始
	// ----------------------------------------------
	var prompt = fmt.Sprintf(" %s ", inputPrompt)
	var inputText string
	//var previousInputText = ""

	// forループ外で宣言する
	var php = wiredrawing.PHPExecuter{
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
	//var counter int = 0
	//var mem runtime.MemStats
	for {
		//counter++
		//func() {
		//	runtime.ReadMemStats(&mem)
		//	fmt.Println(mem.Alloc, mem.TotalAlloc, mem.HeapAlloc, mem.HeapSys)
		//}()
		//php.SetNgFile(ngFile)
		//fmt.Print(config.ColorWrapping("0", prompt))
		if php.IsPermissibleError == true {
			fmt.Print("\033[33m")
			fmt.Print(config.ColorWrapping("33", prompt))
		}
		// 両端のスペースを削除
		var isExit int
		// multiline-ny側で非同期実行するためにphpオブジェクトを渡す
		rawInputText, isExit = wiredrawing.StdInput("", inputText, &php)
		//fmt.Printf("rawInputText: %v :", rawInputText)
		if isExit == 1 {
			continue
		} else if isExit == 2 {
			break
		}

		inputText = strings.TrimRight(strings.TrimSpace(rawInputText), "\n")

		if len(inputText) == 0 {
			runtime.GC()
			continue
		}

		if inputText == "help" {
			fmt.Println(config.ColorWrapping("33", helpMessage))
			continue
		} else if inputText == "error" {
			var wholeErrors = php.WholeErrors()
			for key, value := range wholeErrors {
				fmt.Print(config.ColorWrapping(config.Green, fmt.Sprintf("[%03d] => ", key+1)))
				fmt.Print(config.ColorWrapping(config.Red, value))
			}
			continue
		} else if inputText == "reset errors" {
			php.ResetWholeErrors()
			continue
		} else if inputText == "rollback" {
			php.Rollback()
			isFatal, _ := php.DetectFatalError(1)
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
			cl := php.SetPreviousList(0)
			_, _ = php.Execute(false)
			//logs := php.Cat(1)
			//for index := range logs {
			//	indexStr := fmt.Sprintf("%04d", logs[index]["id"])
			//	fmt.Printf("%v[%s]: ", index+1, config.ColorWrapping("34", indexStr))
			//	var statement string = (logs[index]["text"]).(string)
			//	var _ []string = strings.Split(statement, "\n")
			//	fmt.Println(config.ColorWrapping("32", statement))
			//}
			php.SetPreviousList(cl)
			continue
		}

		if inputText == "save" {
			//saveFileName = "save_" + fmt.Sprintf("%04d", counter) + ".php"
			//php.Save(saveFileName)
			continue
		}
		if inputText == "exit" {
			// コンソールを終了するための標準入力を取得する

			fmt.Println(config.ColorWrapping("31", "[PHPコマンドラインを終了します。本当に終了する場合は<yes>と入力して下さい。]"))
			//fmt.Print(prompt)
			// 両端のスペースを削除
			rawInputText = ""
			var inputText, _ = wiredrawing.StdInput(prompt, rawInputText, &php)
			inputText = strings.TrimSpace(inputText)
			if wiredrawing.InArray(inputText, wordsToExit) {
				// 終了メッセージを表示
				// string型を[]byteに変換して書き込み
				var messageToEnd = []byte("Thank you for using me! Good by.")
				_, err := os.Stdout.Write([]byte(config.ColorWrapping("34", string(messageToEnd))))
				os.Stdout.Write([]byte("\n\n"))
				if err != nil {
					return false, err
				}
				break
			}
			// 入力内容が空文字の場合コマンドラインを終了する
			fmt.Println(config.ColorWrapping("31", "キャンセルしました。"))
			continue
		}

		// 入力内容が [clear] or [refresh] だった場合は入力内容をクリア
		// ファイルサイズを空にする
		if inputText == "clear" {
			//// バリデーション用のファイルを空にする
			//php.Clear()
			//// ただPHP開始タグのみ先頭に追記しておく
			//php.WriteToNg("<?php "+"\n", 1)
			//fmt.Print("\033[0m")
			//php.IsPermissibleError = false
			//prompt = fmt.Sprintf(" %s ", inputPrompt)
			runtime.GC()
			debug.FreeOSMemory()
			continue
		}

		// cat と入力すると現在まで入力している内容を出力する
		if inputText == "cat" {
			logs := php.Cat()
			for index, log := range logs {
				text := log["text"].(string)
				if len(text) == 0 {
					continue
				}
				indexStr := fmt.Sprintf("%04d", index)
				fmt.Print(config.ColorWrapping("34", indexStr) + ": ")
				if len(text) > 0 {
					fmt.Println(config.ColorWrapping("32", text))
				}
			}
			continue
		}

		php.WriteToFile(rawInputText + "\n")
		// ----------------------------------------------------------------
		// Fatal Error以外を検出する
		// ----------------------------------------------------------------
		isFatal, err := php.DetectFatalError(1)
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
					fmt.Println(config.ColorWrapping("31", string(errorBuffer)))
					// 事前検証用のfileの中身を本実行用fileの中身と同じにする
					php.Rollback()
				}
				continue
			}
		} else if (isFatal == false) && len(php.ErrorBuffer) > 0 {
			// Fatal Error ではないがErrorBufferが空ではない場合
			fmt.Println(config.ColorWrapping("33", string(php.ErrorBuffer)))
			// 事前検証用のfileの中身を本実行用fileの中身と同じにする
			php.Rollback()
			continue
		}

		if outputSize, err := php.Execute(true); err != nil {
			fmt.Println(config.ColorWrapping("31", err.Error()))
		} else {
			if outputSize > 0 {
				fmt.Print(config.ColorWrapping("0", "\n"))
			} else {
				fmt.Print(config.ColorWrapping("0", ""))
			}
		}
		inputText = ""
		prompt = fmt.Sprintf(" %s ", inputPrompt)
		continue
	}
	// 標準入力の終了

	return true, nil
}
