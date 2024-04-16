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
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

// 改行文字を定義
const newLine string = "\n"

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
func StandByInput(phpPath string) (bool, error) {
	// phpPathが未指定の場合はハイフンが設定されているため
	// それ以外の場合は指定したphpパスで実行する
	var phpExecutePath = "php"
	if phpPath != "-" {
		phpExecutePath = phpPath
	}
	//var previousLine = new(int)
	//var previousLine = 0
	// PHPのエラーメッセージの正規表現を事前コンパイルする
	//const ParseErrorString = `PHP[ ]+?Parse[ ]+?error:[ ]+?syntax[ ]+?error,`
	//var parseErrorRegex = regexp.MustCompile(ParseErrorString)

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

	//previousLine = 0

	// ----------------------------------------------
	// 標準入力を可能にする
	// 標準入力の開始
	// ----------------------------------------------
	var prompt = " >>> "
	var isPermissibleError = false
	//var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
	//var latestErrorMessage = ""
	var inputText string
	//var previousInputText = ""

	// forループ外で宣言する
	var php = wiredrawing.PhpExecuter{
		PhpPath: phpExecutePath,
	}
	for {
		if isPermissibleError == true {
			fmt.Print("\033[33m")
		}
		fmt.Print(prompt)
		// Terminalのカラーリングを白に戻す
		fmt.Print(colorWrapping("0", ""))
		// 両端のスペースを削除
		rawInputText := wiredrawing.StdInput()
		//previousInputText = inputText
		inputText = strings.TrimSpace(rawInputText)

		if len(inputText) == 0 {
			runtime.GC()
			continue
		}

		if inputText == "rollback" {
			popStirngToFile(ngFile, -1)
			errorBytes, _, _ := php.DetectFatalError()
			if len(errorBytes) > 0 {
				prompt = " ... "
			} else {
				errorBytes, _ := php.DetectErrorExceptFatalError()
				if len(errorBytes) > 0 {
					prompt = " ... "
				} else {
					prompt = " >>> "
				}
			}
			continue
		}

		if inputText == "exit" {
			// コンソールを終了するための標準入力を取得する

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

		if inputText == "ls" {

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
			isPermissibleError = false
			fmt.Print("\033[0m")
			prompt = " >>> "
			runtime.GC()
			debug.FreeOSMemory()
			continue
		}

		// cat と入力すると現在まで入力している内容を出力する
		if inputText == "cat" {
			catFile, err := os.Open(ngFile)
			if err != nil {
				panic(err)
			}
			tempScanner := bufio.NewScanner(catFile)

			var index = 0
			var indexStr = ""
			fmt.Println("")
			for tempScanner.Scan() {
				indexStr = fmt.Sprintf("%03d", index)
				fmt.Print(colorWrapping("34", indexStr) + ": ")
				fmt.Println(colorWrapping("32", tempScanner.Text()))
				index++
			}
			fmt.Println("")
			continue
		}

		php.SetOkFile(okFile)
		php.SetNgFile(ngFile)
		php.WriteToNg(rawInputText + "\n")
		// ----------------------------------------------------------------
		// Fatal Error以外を検出する
		// ----------------------------------------------------------------
		fatalError, isFatal, _ := php.DetectFatalError()
		//fmt.Printf("fatalError ===> %v", fatalError)
		//fmt.Printf("e ===> %v", e)
		if len(fatalError) > 0 {
			if php.IsPermissibleError == true {
				prompt = " ... "
			} else {
				if isFatal == true {
					// Fatal Errorが検出された場合はエラーメッセージを表示して終了
					fmt.Println(colorWrapping("31", string(fatalError)))
				} else {
					// Fatal Errorが検出された場合はエラーメッセージを表示して終了
					fmt.Println(colorWrapping("33", string(fatalError)))
				}
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
		// ----------------------------------------------------------------
		// None Fatal Error一群を検出する
		// ----------------------------------------------------------------
		//noneFatalError, _ := php.DetectErrorExceptFatalError()
		//if len(noneFatalError) > 0 {
		//	{
		//		fmt.Println(colorWrapping("33", string(noneFatalError)))
		//		// 事前検証用のfileの中身を本実行用fileの中身と同じにする
		//		if source, err := os.Open(okFile); err != nil {
		//			panic(err)
		//		} else {
		//			// 現在のOKfileの中身を取得
		//			destination, _ := os.Create(ngFile)
		//			io.Copy(destination, source)
		//		}
		//		continue
		//	}
		//}
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
		if outputSize, err := php.Execute(); err != nil {
			fmt.Println(colorWrapping("31", err.Error()))
		} else {
			if outputSize > 0 {
				fmt.Print(colorWrapping("0", "\n"))
			} else {
				fmt.Print(colorWrapping("0", ""))
			}
		}
		prompt = " >>> "
		continue
		//php.ExecutePhp(previousLine, true)
		//var exitCode int;
		//if (exitCode, err := php.DetectFatalError(); err != nil) {
		//	fmt.Println(err);
		//	continue;
		//} else if (exitCode != 0) {
		//
		//}
		//php.ExecutePhp(previousLine, true)

		// ---------------------------------------------
		// exitコマンドの場合はシェルを終了
		// ---------------------------------------------
		//if inputText == "exit" {
		//	// コンソールを終了するための標準入力を取得する
		//	{
		//		fmt.Println(colorWrapping("31", "PHPコマンドラインを終了します。本当に終了する場合は<yes>と入力して下さい。"))
		//		fmt.Print(prompt)
		//		// 両端のスペースを削除
		//		var inputText = wiredrawing.StdInput()
		//		inputText = strings.TrimSpace(inputText)
		//		// 入力内容が空文字の場合コマンドラインを終了する
		//		if len(inputText) == 0 {
		//			fmt.Println(colorWrapping("31", "キャンセルしました。"))
		//			continue
		//		}
		//
		//		if wiredrawing.InArray(inputText, wordsToExit) {
		//			// 終了メッセージを表示
		//			// string型を[]byteに変換して書き込み
		//			var messageToEnd = []byte("Thank you for using me! Good by.")
		//			_, err := os.Stdout.Write([]byte(colorWrapping("34", string(messageToEnd))))
		//			if err != nil {
		//				return false, err
		//			}
		//			break
		//		}
		//		continue
		//	}
		//}
		//
		//if inputText == "rollback" {
		//	// 入力内容を①行分削除する時
		//	if prompt == " ... " {
		//		// 複数行入力中の場合は<.validation.php>ファイルのみ一行削除する
		//		isDeleted := popStirngToFile(ngFile, -1)
		//		if isDeleted != nil {
		//			panic(isDeleted)
		//		}
		//	} else {
		//		var isDeleted error
		//		isDeleted = popStirngToFile(okFile, -1)
		//		if isDeleted != nil {
		//			panic(isDeleted)
		//		}
		//		isDeleted = popStirngToFile(ngFile, -1)
		//		if isDeleted != nil {
		//			panic(isDeleted)
		//		}
		//	}
		//	if len(previousInputText) > 0 {
		//		fmt.Println(" ... " + colorWrapping("31", " - "+previousInputText))
		//	}
		//	continue
		//}
		//
		//// ---------------------------------------------
		//// スペースで分割して delete indexNumber を取り出す
		//// ---------------------------------------------
		//tokens := strings.Split(inputText, " ")
		//
		//if tokens[0] == "delete" {
		//	var changedPreviousLine = deletePreviousCode(tokens, ngFile, okFile)
		//	// 戻り値が<-1>の場合は,deleteコマンドの入力に不備があった場合
		//	if changedPreviousLine != -1 {
		//		previousLine = changedPreviousLine
		//	}
		//	continue
		//}
		//
		//// 入力内容が [clear] or [refresh] だった場合は入力内容をクリア
		//// ファイルサイズを空にする
		//if inputText == "clear" || inputText == "refresh" {
		//	// phpスクリプトチェック用ファイルを殻にする
		//	_ = os.Truncate(ngFile, 0)
		//	//file1.Seek(0, 0)
		//	f, err := os.Open(okFile)
		//	if err != nil {
		//		panic(err)
		//	}
		//	source, err := ioutil.ReadAll(f)
		//	if err != nil {
		//		panic(err)
		//	}
		//	ioutil.WriteFile(ngFile, source, 0777)
		//	//file1.WriteString(string(source))
		//	isPermissibleError = false
		//	fmt.Print("\033[0m")
		//	prompt = " >>> "
		//	runtime.GC()
		//	debug.FreeOSMemory()
		//	continue
		//}
		//
		//// 入力内容が [show error] だった場合は直前のエラーメッセージを表示
		//if inputText == "show error" {
		//	fmt.Println(colorWrapping("33", latestErrorMessage))
		//	continue
		//}
		//
		//// cat と入力すると現在まで入力している内容を出力する
		//if inputText == "cat" {
		//	catFile, err := os.Open(ngFile)
		//	if err != nil {
		//		panic(err)
		//	}
		//	tempScanner := bufio.NewScanner(catFile)
		//
		//	var index = 0
		//	var indexStr = ""
		//	fmt.Println("")
		//	for tempScanner.Scan() {
		//		indexStr = fmt.Sprintf("%03d", index)
		//		fmt.Print(colorWrapping("34", indexStr) + ": ")
		//		fmt.Println(colorWrapping("32", tempScanner.Text()))
		//		index++
		//	}
		//	fmt.Println("")
		//	continue
		//}
		//
		//// プログラムの実行処理
		//// ファイルポインタへ書き込み
		//if _, err := wiredrawing.FileOpen(ngFile, rawInputText+newLine); err != nil {
		//	return false, err
		//}
		//
		//// --------------------------------------------
		//// 一旦入力内容が正しく終了するかどうかを検証
		//// --------------------------------------------
		////command := exec.Command("php", ngFile)
		////
		////var e = command.Run()
		////if e != nil {
		////	// PHPコマンドの実行に失敗した場合
		////}
		////exitCode := command.ProcessState.ExitCode()
		//exitCode := wiredrawing.ValidateVNgFile(ngFile)
		//// エラーコードが0でない場合
		//if exitCode != 0 {
		//	var _ *os.File
		//	command := exec.Command("php", ngFile)
		//	buffer, err := command.StderrPipe()
		//	if err != nil {
		//		//log.Fatal(err)
		//	}
		//
		//	if err := command.Start(); err != nil {
		//		//log.Fatal(err)
		//	}
		//	slurp, _ := io.ReadAll(buffer)
		//	if err := command.Wait(); err != nil {
		//		//log.Fatal(err)
		//	}
		//	latestErrorMessage = string(slurp)
		//
		//	isPermissibleError = false
		//	if parseErrorRegex.MatchString(string(slurp)) {
		//		isPermissibleError = true
		//	}
		//	//for _, value := range compiledPHPErrorMessages {
		//	//	// 規定したエラーメッセージにマッチした場合はokFileの中身をngFileにコピーする
		//	//	if value.MatchString(string(slurp)) {
		//	//
		//	//		break
		//	//	}
		//	//}
		//	// 許容可能なエラーでは無い場合
		//	if isPermissibleError == false {
		//		os.Stdout.Write([]byte(colorWrapping("31", string(slurp))))
		//		// phpスクリプトチェック用ファイルを殻にする
		//		_ = os.Truncate(ngFile, 0)
		//		//file1.Seek(0, 0)
		//		f, err := os.Open(okFile)
		//		if err != nil {
		//			panic(err)
		//		}
		//		source, err := ioutil.ReadAll(f)
		//		if err != nil {
		//			panic(err)
		//		}
		//		ioutil.WriteFile(ngFile, source, 0777)
		//		//file1.WriteString(string(source))
		//		isPermissibleError = false
		//		prompt = " >>> "
		//		continue
		//	}
		//
		//	// 継続可能な許容エラーの場合
		//	isPermissibleError = true
		//	//fmt.Print(colorWrapping("37", "\tERROR: "+string(slurp)))
		//	fmt.Fprint(os.Stdout, "\033[33m")
		//	prompt = " ... "
		//	continue
		//}
		//isPermissibleError = false
		//
		//// NoticeやWarningなどのエラーを検出する
		//command := exec.Command("php", ngFile)
		//noneFatalErrorBuffer, _ := command.StderrPipe()
		//_ = command.Start()
		//success, _ := io.ReadAll(noneFatalErrorBuffer)
		//if err := command.Wait(); err != nil {
		//	log.Fatal(err)
		//}
		//if len(success) > 0 {
		//	// 単純に標準エラーに何かしら出力されていたらエラーとする
		//	_ = os.Truncate(ngFile, 0)
		//	f, err := os.Open(okFile)
		//	if err != nil {
		//		//panic(err)
		//	}
		//	source, err := ioutil.ReadAll(f)
		//	if err != nil {
		//		//panic(err)
		//	}
		//	_ = ioutil.WriteFile(ngFile, source, 0777)
		//	fmt.Print(colorWrapping("33", string(success)))
		//	continue
		//}
		//prompt = " >>> "
		//
		//// --------------------------------------------
		//// 正常終了の場合は ngFile中身をokFileにコピー
		//// --------------------------------------------
		//ngf, err := os.Open(ngFile)
		//okf, err := os.Create(okFile)
		//_, err = io.Copy(okf, ngf)
		////fmt.Printf("書き込まれたsize: %d\n", size)
		//ngf.Close()
		//okf.Close()
		//if err != nil {
		//	panic(err)
		//}
		////if err != nil {
		////	panic(err)
		////}
		////b, err := ioutil.ReadAll(f)
		////if err != nil {
		////	panic(err)
		////}
		////os.Truncate(okFile, 0)
		////ioutil.WriteFile(okFile, b, 0777)
		//
		//command = exec.Command(phpExecutePath, okFile)
		//buffer, err := command.StdoutPipe()
		//
		//if err != nil {
		//	panic(err)
		//}
		//// bufferの読み取り開始
		//err = command.Start()
		//if err != nil {
		//	panic(err)
		//}
		//// 第三引数にtrueを与えて出力結果を表示する
		//// 出力文字列の色を青にして出力
		//_, outputSize := wiredrawing.LoadBuffer(buffer, &previousLine, true, false, "34")
		//err = command.Wait()
		//if (err != nil) && (err.Error() != "exit status 255") {
		//	panic(err)
		//}
		////fmt.Printf("outputSize: %d\n", outputSize)
		//if outputSize > 0 {
		//	fmt.Print(colorWrapping("0", "\n"))
		//} else {
		//	fmt.Print(colorWrapping("0", ""))
		//}
		//latestErrorMessage = ""
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
		wiredrawing.LoadBuffer(buffer, &previousLine, false, false, "34")
		fmt.Println("")
		return -1
	}
	// 正常に動作した最終のバイト数を返却する
	return previousLine
}
