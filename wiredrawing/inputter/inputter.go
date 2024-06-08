package inputter

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"phpgo/wiredrawing"
	"runtime"
	"runtime/debug"
	"strconv"
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
	var dotDir = homeDir + "\\.php-go"
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
func StandByInput(phpPath string, inputPrompt string, saveFileName string, exit chan int) (bool, error) {
	// phpPathが未指定の場合はハイフンが設定されているため
	// それ以外の場合は指定したphpパスで実行する
	var phpExecutePath = "php"
	if phpPath != "-" {
		phpExecutePath = phpPath
	}
	// ターミナルを終了させるためのキーワード群
	var wordsToExit = make([]string, 0, 32)

	// ターミナル終了キーワードを設定
	wordsToExit = append(wordsToExit, "y")
	wordsToExit = append(wordsToExit, "Y")
	wordsToExit = append(wordsToExit, "yes")

	var ngFile = ".validation.dat"
	var okFile = ".success.dat"
	var filePathForError = ".erorr_message.dat"

	var homeDir string
	homeDir, _ = os.UserHomeDir()
	// 本アプリケーション専用の設定ディレクトリ
	var dotDir = homeDir + "\\.php-go"
	// ディレクトリが存在しない場合は作成する
	makeDirectory(dotDir)

	ngFile = dotDir + "\\" + ngFile
	okFile = dotDir + "\\" + okFile
	filePathForError = dotDir + "\\" + filePathForError

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

	//previousLine = 0

	// ----------------------------------------------
	// 標準入力を可能にする
	// 標準入力の開始
	// ----------------------------------------------
	var prompt = fmt.Sprintf(" %s ", inputPrompt)
	//var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
	//var latestErrorMessage = ""
	var inputText string
	//var previousInputText = ""

	// forループ外で宣言する
	var php = wiredrawing.PhpExecuter{
		PhpPath: phpExecutePath,
	}
	for {
		if php.IsPermissibleError == true {
			fmt.Print("\033[33m")
		}
		fmt.Print(prompt)
		// Terminalのカラーリングを白に戻す
		fmt.Print(ColorWrapping("0", ""))
		// 両端のスペースを削除
		rawInputText := wiredrawing.StdInput()
		//previousInputText = inputText
		inputText = strings.TrimSpace(rawInputText)

		if len(inputText) == 0 {
			runtime.GC()
			continue
		}

		if inputText == "help" {
			var helpMessages = make([]string, 0, 32)
			helpMessages = append(helpMessages, fmt.Sprintf("[help]"))
			helpMessages = append(helpMessages, fmt.Sprintf("clear:    入力途中の内容を破棄します."))
			helpMessages = append(helpMessages, fmt.Sprintf("rollback: 入力済みの入力を一行ずつ削除します."))
			helpMessages = append(helpMessages, fmt.Sprintf("cat:      入力済みの内容を表示します."))
			helpMessages = append(helpMessages, fmt.Sprintf("exit:     アプリケーションを終了します."))
			helpMessages = append(helpMessages, fmt.Sprintf("errors:   過去のエラーを全て表示します."))
			var helpMessage = strings.Join(helpMessages, "\n")
			fmt.Println(ColorWrapping("33", helpMessage))
			continue
		}

		if inputText == "errors" {
			var wholeErrors = php.WholeErrors()
			for key, value := range wholeErrors {
				fmt.Print(ColorWrapping(Green, fmt.Sprintf("[%03d] => ", key+1)))
				fmt.Print(ColorWrapping(Red, value))
			}
			continue
		}
		if inputText == "reset errors" {
			php.ResetWholeErrors()
			continue
		}
		if inputText == "rollback" {
			lines, err := wiredrawing.File(ngFile)
			if err != nil {
				panic(err)
			}
			if len(lines) == 1 {
				continue
			}
			_ = popStirngToFile(ngFile, -1)
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
			php.SetOkFile(ngFile)
			php.SetPreviousList(0)
			php.Execute(false)
			_ = concatenate(ngFile)
			continue
		}

		if inputText == "save" {
			php.Save(saveFileName)
			continue
		}
		if inputText == "exit" {
			// コンソールを終了するための標準入力を取得する

			fmt.Println(ColorWrapping("31", "[PHPコマンドラインを終了します。本当に終了する場合は<yes>と入力して下さい。]"))
			fmt.Print(prompt)
			// 両端のスペースを削除
			var inputText = wiredrawing.StdInput()
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
			ng, err := os.Create(ngFile)
			ok, err := os.Open(okFile)
			writtenSize, err := io.Copy(ng, ok)
			// 出力を削除
			_, _ = fmt.Fprintln(ioutil.Discard, writtenSize, err)
			fmt.Print("\033[0m")
			php.IsPermissibleError = false
			prompt = fmt.Sprintf(" %s ", inputPrompt)
			runtime.GC()
			debug.FreeOSMemory()
			continue
		}

		// cat と入力すると現在まで入力している内容を出力する
		if inputText == "cat" {
			concatenate(ngFile)
			continue
		}

		php.SetOkFile(okFile)
		php.SetNgFile(ngFile)
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
					if source, err := os.Open(okFile); err != nil {
						panic(err)
					} else {
						if destination, err := os.Create(ngFile); err != nil {
							panic(err)
						} else {
							size, err := io.Copy(destination, source)
							_, _ = fmt.Fprintln(ioutil.Discard, size, err)
						}
					}
				}
				continue
			}
		} else if (isFatal == false) && len(php.ErrorBuffer) > 0 {
			// Fatal Error ではないがErrorBufferが空ではない場合
			fmt.Println(ColorWrapping("33", string(php.ErrorBuffer)))
			// 事前検証用のfileの中身を本実行用fileの中身と同じにする
			if source, err := os.Open(okFile); err != nil {
				panic(err)
			} else {
				if destination, err := os.Create(ngFile); err != nil {
					panic(err)
				} else {
					size, err := io.Copy(destination, source)
					_, _ = fmt.Fprintln(ioutil.Discard, size, err)
				}
			}
			continue
		}

		// Fatalエラーそれ以外のエラーともに検出されなかった場合
		// ngFileの中身をokFileにコピー
		if ok, err := os.Create(okFile); err != nil {
			panic(err)
		} else {
			if ng, err := os.Open(ngFile); err != nil {
				panic(err)
			} else {
				_, err = io.Copy(ok, ng)
				if err != nil {
					panic(err)
				}
			}
		}
		if outputSize, err := php.Execute(true); err != nil {
			fmt.Println(ColorWrapping("31", err.Error()))
		} else {
			if outputSize > 0 {
				fmt.Print(ColorWrapping("0", "\n"))
			} else {
				fmt.Print(ColorWrapping("0", ""))
			}
		}
		prompt = fmt.Sprintf(" %s ", inputPrompt)
		continue
	}
	// 標準入力の終了

	return true, nil
}

func hasIndex(slice []string, index int) bool {
	return (0 <= index) && (index < len(slice))
}

func ColorWrapping(colorCode string, text string) string {
	return "\033[" + colorCode + "m" + text + "\033[0m"
}

// popStirngToFile 指定したファイルの指定した行を削除する
// 指定する行数は 1から開始させる
func popStirngToFile(filePath string, row int) error {
	// 削除する行の指定は -1 あるいは 1以上とする
	// -1の場合は最後の行を削除する
	if row != -1 && row <= 0 {
		return fmt.Errorf("the row number is invalid")
	}
	var file *os.File
	file, err := os.OpenFile(filePath, os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	var bodyByte []byte
	// 一行ずつ読み込んだ結果の文字列の配列
	var bodyString []string
	var temp []byte
	bodyByte, err = ioutil.ReadAll(file)
	for _, value := range bodyByte {
		if value == '\n' {
			s := string(temp)
			if len(s) > 0 {
				bodyString = append(bodyString, s)
				temp = nil
			}
			continue
		}
		temp = append(temp, value)
	}
	if len(bodyString) == 1 {
		// つまり <?php のみの場合
		return nil
	}
	// 削除する行の指定が -1 の場合は最後の行を削除する
	//fmt.Printf("%v", bodyString)
	if row == -1 {
		bodyString = bodyString[:len(bodyString)-1]
	} else {
		bodyString = append(bodyString[:row-1], bodyString[row:]...)
	}
	//fmt.Printf("%v", bodyString)
	// 削除した結果の[]string変数が => bodyString に格納されている
	os.Truncate(filePath, 0)
	_, err = file.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	//bodyString = append(bodyString, "\n")
	connectedBodyString := strings.Join(bodyString, "\n")
	//fmt.Println("connectedBodyString: ", connectedBodyString)
	_, err = file.WriteString(connectedBodyString)
	if err != nil {
		panic(err)
	}
	//fmt.Println("writtenSize: ", writtenSize)
	file.WriteString("\n")
	file.Close()
	return nil
}

// deletePreviousCode 直前に入力したコードを削除する
func deletePreviousCode(tokens []string, ngFile string, okFile string) int {
	var previousLine = 0
	{
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
		// 削除したい対象の行数を取得する
		var indexToDelete int
		if hasIndex(tokens, 1) != true {
			fmt.Println("削除したい行数を指定してください")
			return -1
		}
		indexToDelete, err = strconv.Atoi(tokens[1])
		if err != nil {
			panic(err)
		}

		// 指定したindexがスライスの範囲内かどうかを検証
		if (len(fileBuffer) > indexToDelete) != true {
			fmt.Println("範囲外のindexが指定されました")
			return -1
		}
		fileBuffer = append(fileBuffer[:indexToDelete], fileBuffer[indexToDelete+1:]...)
		var closeError = file.Close()
		if closeError != nil {
			panic(closeError)
		}

		// validation用ファイルを空に
		//file1.Close()
		_ = os.Truncate(ngFile, 0)
		file, err = os.OpenFile(ngFile, os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			panic(err)
		}
		_, _ = file.Seek(0, 0)
		for _, value := range fileBuffer {
			// 行末に改行文字を入力
			_, err := file.WriteString(value + "\n")
			if err != nil {
				panic(err)
			}
		}
		_ = file.Close()

		_ = os.Truncate(okFile, 0)
		file, err = os.OpenFile(okFile, os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			panic(err)
		}
		_, _ = file.Seek(0, 0)
		for _, value := range fileBuffer {
			// 行末に改行文字を入力
			_, err := file.WriteString(value + "\n")
			if err != nil {
				panic(err)
			}
		}
		_ = file.Close()

		// 再度削除したphpファイルを実行して古いバッファを捨てる
		command := exec.Command("php", okFile)
		buffer, err := command.StdoutPipe()
		if err != nil {
			panic(err)
		}
		command.Start()
		// 第三引数にfalseを与えて,実行結果の出力を破棄する
		wiredrawing.LoadBuffer(buffer, &previousLine, false, false, Blue)
		fmt.Println("")
		return -1
	}
	// 正常に動作した最終のバイト数を返却する
	return previousLine
}
