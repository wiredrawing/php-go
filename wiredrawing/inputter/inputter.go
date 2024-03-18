package inputter

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"php-go/wiredrawing"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

// 改行文字を定義
const newLine string = "\n"

// 入力内容を保持する変数
var inputText = ""

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
func StandByInput() (bool, error) {
	var previousLine = new(int)
	// PHPのエラーメッセージの正規表現を事前コンパイルする
	var PHPErrorMesssages []string = []string{
		`Warning:[ ]+?Uncaught[ ]+?Error:[ ]+?Closure[ ]+?object[ ]+?cannot[ ]+?have[ ]+?propecatrties`,
		`Fatal[ ]+?error:[ ]+?Uncaught[ ]+?ArgumentCountError:[ ]+?Too[ ]+?few[ ]+?arguments[ ]+?to[ ]+?function[ ]+?(.)*?::(.)*?`,
		`Fatal[ ]+?error:[ ]+?Uncaught[ ]+?Error:[ ]+?Closure[ ]+?object[ ]+?cannot[ ]+?have[ ]+?properties`,
		`Fatal[ ]+?error:[ ]+?Call[ ]+?to[ ]+?undefined[ ]+?function[ ]+?(.+)?\(\) in`,
		`Fatal[ ]+?error:[ ]+?Uncaught[ ]+?Error:[ ]+?Call[ ]+to[ ]+?undefined[ ]+?function[ ]+?.+?`,
		`Fatal[ ]+?error:[ ]+?Uncaught[ ]+?Error:[ ]+?Class[ ]+'.+?'[ ]+not[ ]+found[ ]+in`,
		`Fatal[ ]+?error:[ ]+?Uncaught[ ]+?ArgumentCountError:[ ]+?Too[ ]+?few[ ]+?arguments[ ]+?to[ ]+?function[ ]+?{closure}\(\),[ ]+?([0-9]+?)[ ]+?passed in`,
	}
	var compiledPHPErrorMessages = make([]*regexp.Regexp, 0, 2)
	for _, value := range PHPErrorMesssages {
		var r *regexp.Regexp = regexp.MustCompile(value)
		compiledPHPErrorMessages = append(compiledPHPErrorMessages, r)
	}
	// PHPの Fatal Error 以外のエラーメッセージの正規表現を事前コンパイルする
	var PHPNoneErrorMessages []string = []string{
		`Warning:[ ]+?Use of undefined constant (.+)? - assumed '(.+)?' (this will throw an Error in a future version of PHP)`,
		`Warning:[ ]+?(.+)?\(\)[ ]+?expects[ ]+?at[ ]+?least[ ]+?[0-9]+?[ ]+?parameter(s)?,[ ]+?[0-9]+?[ ]+?given[ ]+?in`,
		`Warning:[ ]+?Use[ ]+?of[ ]+?undefined[ ]+?constant[ ]+?(.+)?[ ]+?\-[ ]+?assumed[ ]+?'(.+)?'[ ]+?\(this will throw an Error in a future version of PHP\)[ ]+?in`,
		`Notice:[ ]+?Undefined index: .+?`,
		`Notice:[ ]+?Undefined variable: .+? in (.+)? on line (.+)?`,
		`Deprecated:[ ]+?Methods with the same name as their class will not be constructors in a future version of PHP; (.+)? has a deprecated constructor`,
	}
	var compiledPHPNoneErrorMessages = make([]*regexp.Regexp, 0, 2)
	for _, value := range PHPNoneErrorMessages {
		var r *regexp.Regexp = regexp.MustCompile(value)
		compiledPHPNoneErrorMessages = append(compiledPHPNoneErrorMessages, r)
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

	*previousLine = 0

	// ----------------------------------------------
	// 標準入力を可能にする
	// 標準入力の開始
	// ----------------------------------------------
	var prompt = " >>> "
	var isPermissibleError = false
	//var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
	var latestErrorMessage = ""
	for {
		if isPermissibleError == true {
			fmt.Print("\033[33m")
		}
		fmt.Print(prompt)
		// Terminalのカラーリングを白に戻す
		fmt.Print(colorWrapping("0", ""))
		// 両端のスペースを削除
		rawInputText := wiredrawing.StdInput()
		inputText = strings.TrimSpace(rawInputText)
		// 入力内容が exit ならアプリケーションを終了
		if len(inputText) == 0 {
			runtime.GC()
			//debug.FreeOSMemory()
			continue
		}

		// ---------------------------------------------
		// exitコマンドの場合はシェルを終了
		// ---------------------------------------------
		if inputText == "exit" {
			// コンソールを終了するための標準入力を取得する
			{
				fmt.Println(colorWrapping("31", "PHPコマンドラインを終了します。本当に終了する場合は<yes>と入力して下さい。"))
				fmt.Print(prompt)
				// 両端のスペースを削除
				var inputText = wiredrawing.StdInput()
				inputText = strings.TrimSpace(inputText)
				// 入力内容が空文字の場合コマンドラインを終了する
				if len(inputText) == 0 {
					fmt.Println(colorWrapping("31", "キャンセルしました。"))
					continue
				}

				if wiredrawing.InArray(inputText, wordsToExit) {
					// 終了メッセージを表示
					// string型を[]byteに変換して書き込み
					var messageToEnd = []byte("Thank you for using me! Good by.")
					_, err := os.Stdout.Write([]byte(colorWrapping("34", string(messageToEnd))))
					if err != nil {
						return false, err
					}
					break
				}
				continue
			}
		}

		if inputText == "rollback" {
			// 入力内容を①行分削除する時
			if prompt == " ... " {
				// 複数行入力中の場合は<.validation.php>ファイルのみ一行削除する
				isDeleted := popStirngToFile(ngFile, -1)
				if isDeleted != nil {
					panic(isDeleted)
				}
				//fp, err := os.Open(ngFile)
				//if err != nil {
				//	panic(err)
				//}
				//s := bufio.NewScanner(fp)
				//var rollbackData []string
				//for s.Scan() {
				//	rollbackData = append(rollbackData, s.Text())
				//}
				//replacedRollbackData := rollbackData[:len(rollbackData)-1]
				//fp.Close()
				//if len(replacedRollbackData) > 1 {
				//	os.Truncate(ngFile, 0)
				//	fp, _ = os.OpenFile(ngFile, os.O_APPEND|os.O_WRONLY, 0777)
				//	for _, value := range replacedRollbackData {
				//		fp.WriteString(value + "\n")
				//	}
				//	fp.Close()
				//}
			} else {
				var isDeleted error
				isDeleted = popStirngToFile(okFile, -1)
				if isDeleted != nil {
					panic(isDeleted)
				}
				isDeleted = popStirngToFile(ngFile, -1)
				if isDeleted != nil {
					panic(isDeleted)
				}
			}
			continue
		}

		// ---------------------------------------------
		// スペースで分割して delete indexNumber を取り出す
		// ---------------------------------------------
		tokens := strings.Split(inputText, " ")

		if tokens[0] == "delete" {
			{
				//file1.Close()
				//file2.Close()
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
					continue
				}
				indexToDelete, err = strconv.Atoi(tokens[1])
				if err != nil {
					panic(err)
				}

				// 指定したindexがスライスの範囲内かどうかを検証
				if (len(fileBuffer) > indexToDelete) != true {
					fmt.Println("範囲外のindexが指定されました")
					continue
				}
				fileBuffer = append(fileBuffer[:indexToDelete], fileBuffer[indexToDelete+1:]...)
				var closeError = file.Close()
				if closeError != nil {
					panic(closeError)
				}

				// validation用ファイルを空に
				//file1.Close()
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
				//file2.Close()
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

				// 再度削除したphpファイルを実行して古いバッファを捨てる
				command := exec.Command("php", okFile)
				buffer, err := command.StdoutPipe()
				if err != nil {
					panic(err)
				}
				command.Start()
				*previousLine = 0
				// 第三引数にfalseを与えて,実行結果の出力を破棄する
				wiredrawing.LoadBuffer(buffer, previousLine, false, false, "34")
				fmt.Println("")
				continue
			}

		}

		// 入力内容が [clear] or [refresh] だった場合は入力内容をクリア
		// ファイルサイズを空にする
		if inputText == "clear" || inputText == "refresh" {
			// phpスクリプトチェック用ファイルを殻にする
			_ = os.Truncate(ngFile, 0)
			//file1.Seek(0, 0)
			f, err := os.Open(okFile)
			if err != nil {
				panic(err)
			}
			source, err := ioutil.ReadAll(f)
			if err != nil {
				panic(err)
			}
			ioutil.WriteFile(ngFile, source, 0777)
			//file1.WriteString(string(source))
			isPermissibleError = false
			fmt.Print("\033[0m")
			prompt = " >>> "
			runtime.GC()
			debug.FreeOSMemory()
			continue
		}

		// 入力内容が [show error] だった場合は直前のエラーメッセージを表示
		if inputText == "show error" {
			fmt.Println(colorWrapping("33", latestErrorMessage))
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

				var index = 0
				var indexStr = ""
				for tempScanner.Scan() {
					indexStr = fmt.Sprintf("%03d", index)
					fmt.Print(colorWrapping("34", indexStr) + ": ")
					fmt.Println(colorWrapping("32", tempScanner.Text()))
					index++
				}
			})()

			continue
		}

		// プログラムの実行処理
		// ファイルポインタへ書き込み
		_, err := wiredrawing.FileOpen(ngFile, rawInputText+newLine)
		if err != nil {
			return false, err
		}

		// --------------------------------------------
		// 一旦入力内容が正しく終了するかどうかを検証
		// --------------------------------------------
		command := exec.Command("php", ngFile)

		var e = command.Run()
		if e != nil {
			// PHPコマンドの実行に失敗した場合
		}
		exitCode := command.ProcessState.ExitCode()
		// エラーコードが0でない場合
		if exitCode != 0 {
			var _ *os.File
			command := exec.Command("php", ngFile)
			buffer, err := command.StderrPipe()
			if err != nil {
				//log.Fatal(err)
			}

			if err := command.Start(); err != nil {
				//log.Fatal(err)
			}
			slurp, _ := io.ReadAll(buffer)
			if err := command.Wait(); err != nil {
				//log.Fatal(err)
			}
			latestErrorMessage = string(slurp)

			isPermissibleError = true
			for _, value := range compiledPHPErrorMessages {
				// 規定したエラーメッセージにマッチした場合はokFileの中身をngFileにコピーする
				if value.MatchString(string(slurp)) {
					// phpスクリプトチェック用ファイルを殻にする
					_ = os.Truncate(ngFile, 0)
					//file1.Seek(0, 0)
					f, err := os.Open(okFile)
					if err != nil {
						panic(err)
					}
					source, err := ioutil.ReadAll(f)
					if err != nil {
						panic(err)
					}
					ioutil.WriteFile(ngFile, source, 0777)
					//file1.WriteString(string(source))
					isPermissibleError = false
					break
				}
			}
			// 許容可能なエラーでは無い場合
			if isPermissibleError == false {
				os.Stdout.Write([]byte(colorWrapping("31", string(slurp))))
				prompt = " >>> "
				continue
			}

			// 継続可能な許容エラーの場合
			isPermissibleError = true
			fmt.Fprint(os.Stdout, "\033[33m")
			prompt = " ... "
			continue
		}
		isPermissibleError = false

		// NoticeやWarningなどのエラーを検出する
		command = exec.Command("php", ngFile)
		noneFatalErrorBuffer, _ := command.StderrPipe()
		_ = command.Start()
		success, _ := io.ReadAll(noneFatalErrorBuffer)
		//fmt.Printf("...%v,,,,", string(success))
		if err := command.Wait(); err != nil {
			log.Fatal(err)
		}
		for _, value := range compiledPHPNoneErrorMessages {
			if value.MatchString(string(success)) {
				// 非Fatalエラーにマッチしたエラーの場合はロールバックさせる
				// phpスクリプトチェック用ファイルを殻にする
				_ = os.Truncate(ngFile, 0)
				f, err := os.Open(okFile)
				if err != nil {
					panic(err)
				}
				source, err := ioutil.ReadAll(f)
				if err != nil {
					panic(err)
				}
				_ = ioutil.WriteFile(ngFile, source, 0777)
				fmt.Print(colorWrapping("33", string(success)))
				break
			}
		}

		prompt = " >>> "

		// --------------------------------------------
		// 正常終了の場合は ngFile中身をokFileにコピー
		// --------------------------------------------
		f, err := os.Open(ngFile)
		if err != nil {
			panic(err)
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}
		os.Truncate(okFile, 0)
		ioutil.WriteFile(okFile, b, 0777)
		command = exec.Command("php", okFile)
		buffer, err := command.StdoutPipe()

		if err != nil {
			panic(err)
		}
		// bufferの読み取り開始
		err = command.Start()
		if err != nil {
			panic(err)
		}
		// 第三引数にtrueを与えて出力結果を表示する
		// 出力文字列の色を青にして出力
		_, outputSize := wiredrawing.LoadBuffer(buffer, previousLine, true, false, "34")
		err = command.Wait()
		if (err != nil) && (err.Error() != "exit status 255") {
			panic(err)
		}
		//fmt.Printf("outputSize: %d\n", outputSize)
		if outputSize > 0 {
			fmt.Print(colorWrapping("0", "\n"))
		} else {
			fmt.Print(colorWrapping("0", ""))
		}

	}
	// 標準入力の終了

	return true, nil
}

func hasIndex(slice []string, index int) bool {
	return (0 <= index) && (index < len(slice))
}

func colorWrapping(colorCode string, text string) string {
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
	fmt.Printf("%v", bodyString)
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
