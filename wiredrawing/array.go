package wiredrawing

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"unsafe"
)

// InArray ------------------------------------------------
// PHPのin_array関数をシミュレーション
// 第二引数 haystackに第一引数 needleが含まれていれば
// true それ以外は false
// ------------------------------------------------
func InArray(needle string, haystack []string) bool {

	// 第二引数に指定されたスライスをループさせる
	for _, value := range haystack {
		if needle == value {
			return true
		}
	}

	return false
}

// ArraySearch ------------------------------------------------
// PHPのarray_search関数をシミュレーション
// 第一引数にマッチする要素のキーを返却
// 要素が対象のスライス内に存在しない場合は-1
// ------------------------------------------------
func ArraySearch(needle string, haystack []string) int {

	for index, value := range haystack {
		if value == needle {
			return index
		}
	}
	return -1
}

// StdInput ----------------------------------------
// 標準入力から入力された内容を文字列で返却する
// ----------------------------------------
func StdInput() string {

	// 入力モードの選択用入力
	var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
	var ressult bool = scanner.Scan()
	if ressult == true {
		var which string = scanner.Text()
		if which == ">>>" {
			fmt.Print("\033[33m")
			var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
			var readString []string
			var s = 0
			for {
				if s > 0 {
					fmt.Printf("%s%s", " ... ", " ... ")
				} else {
					fmt.Printf("%s%s", " ... ", " >>> ")
				}
				scanner.Scan()
				value := scanner.Text()

				if value == "rollback" {
					if len(readString) > 0 {
						var lastString string = readString[len(readString)-1]
						readString = readString[0 : len(readString)-1]
						fmt.Print("\v" + colorWrapping("31", lastString) + "\n")
						continue
					}
				} else if value == "cat" {
					// 現在までの入力を確認する
					var indexStr string = ""
					for index, value := range readString {
						indexStr = fmt.Sprintf("%03d", index)
						fmt.Print(colorWrapping("34", indexStr) + ": ")
						fmt.Println(colorWrapping("32", value))
					}
					continue
				}
				if value == "" {
					break
				}
				readString = append(readString, value)
				s++
				//fmt.Printf("k: %v, v: %v\n", key, key)
			}
			fmt.Print("\033[0m")
			if len(readString) > 0 {
				return strings.Join(readString, "\n")
			}
			return ""
		} else {
			return which
		}
	}
	//fmt.Println("Failed scanner.Scan().")
	return ""
}

func colorWrapping(colorCode string, text string) string {
	return "\033[" + colorCode + "m" + text + "\033[0m"
}

type PhpExecuter struct {
	PhpPath            string
	IsPermissibleError bool
	ErrorBuffer        []byte
	SuccessBuffer      []byte
	okFile             string
	ngFile             string
	previousLine       int
}

// SetPreviousList ----------------------------------------
// 前回のセーブポイントを変更する
func (pe *PhpExecuter) SetPreviousList(number int) {
	pe.previousLine = number
}
func (pe *PhpExecuter) SetPhpExcutePath(phpPath string) {
	if phpPath == "" {
		pe.PhpPath = "php"
	}
	pe.PhpPath = phpPath
}

func (pe *PhpExecuter) Execute() (int, error) {
	var showBuffer bool = true
	var colorCode string = "34"
	// isValidate == true の場合はngFileを実行(事前実行)
	command := exec.Command(pe.PhpPath, pe.okFile)

	buffer, err := command.StdoutPipe()
	if err != nil {
		return 0, err
	}
	err = command.Start()
	if err != nil {
		return 0, err
	}
	var currentLine int

	const ensureLength int = 128

	currentLine = 0
	var outputSize int = 0
	// whenError == true の場合バッファ内容を返却してやる
	//var bufferWhenError string
	_, _ = os.Stdout.WriteString("\033[" + colorCode + "m")
	for {
		readData := make([]byte, ensureLength)
		n, err := buffer.Read(readData)
		if (err != nil) && (err != io.EOF) {
			os.Stderr.Write([]byte(err.Error()))
			break
		}
		if n == 0 {
			break
		}
		// 正味のバッファを取り出す
		readData = readData[:n]
		//bufferWhenError += string(readData)

		from := currentLine
		to := currentLine + n
		if (currentLine + n) >= pe.previousLine {
			if from < pe.previousLine && pe.previousLine <= to {
				diff := pe.previousLine - currentLine
				tempSlice := readData[diff:]
				// 出力内容の表示フラグがtrueの場合のみ
				if showBuffer == true {
					outputSize += len(tempSlice)
					_, err = os.Stdout.WriteString(*(*string)(unsafe.Pointer(&tempSlice)))
					if err != nil {
						log.Fatal(err)
					}
				}
			} else {
				// 出力内容の表示フラグがtrueの場合のみ
				if showBuffer == true {
					outputSize += len(readData)
					_, err = os.Stdout.WriteString(*(*string)(unsafe.Pointer(&readData)))
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}
		currentLine += n
		readData = nil
	}
	pe.previousLine = currentLine
	_ = command.Wait()
	// 使用したメモリを開放してみる
	runtime.GC()
	debug.FreeOSMemory()
	// コンソールのカラーをもとにもどす
	_, _ = os.Stdout.WriteString("\033[0m")
	//debug.FreeOSMemory()
	return outputSize, nil
}

// DetectFatalError ----------------------------------------
// 事前にPHPの実行結果がエラーであるかどうかを判定する
func (pe *PhpExecuter) DetectFatalError() ([]byte, bool, error) {

	defer func() {
		if err := recover(); err != nil {
			if err, ok := err.(error); ok {
				fmt.Println(err)
				return
				//return []byte{}, err
			}
		}
		return
		//return []byte{}, errors.New("failed to detect fatal error")
	}()
	//panic(errors.New("意図しないエラー"))
	// PHPのエラーメッセージの正規表現を事前コンパイルする
	const ParseErrorString = `PHP[ ]+?Parse[ ]+?error:[ ]+?syntax[ ]+?error,`
	var parseErrorRegex = regexp.MustCompile(ParseErrorString)
	pe.IsPermissibleError = false
	c := exec.Command(pe.PhpPath, pe.ngFile)
	// コマンド実行失敗時
	if err := c.Run(); err != nil {
		// throw
	}

	// 終了コードが不正な場合,FatalErrorを取得する
	c = exec.Command(pe.PhpPath, pe.ngFile)
	buffer, _ := c.StderrPipe()
	_ = c.Start()
	loadedByte, err := io.ReadAll(buffer)
	if err != nil {
		// 実行結果の出力, PHPのFatal Errorかどうか, Goのエラー
		return []byte{}, true, err
	}
	_ = c.Wait()

	if c.ProcessState.ExitCode() == 0 {
		if len(loadedByte) > 0 {
			// Fatal Error in PHP ではない
			return loadedByte, false, nil
		}
		// 終了コードが正常な場合,何もしない
		return []byte{}, false, nil
	}
	//loadedByte, _ := ioutil.ReadAll(buffer)
	// エラー内容がシンタックスエラーなら許容する
	if parseErrorRegex.MatchString(string(loadedByte)) {
		pe.IsPermissibleError = true
	}
	// シンタックスエラーのみ許容するが Fatal Error in PHP である
	return loadedByte, true, nil
}

func (pe *PhpExecuter) DetectErrorExceptFatalError() ([]byte, error) {
	c := exec.Command(pe.PhpPath, pe.ngFile)
	buffer, err := c.StderrPipe()
	if err != nil {
		return []byte{}, err
	}
	_ = c.Start()
	loadedByte, err := ioutil.ReadAll(buffer)
	_ = c.Wait()
	return loadedByte, nil
}

func (pe *PhpExecuter) GetFatalError() []byte {
	c := exec.Command(pe.PhpPath, pe.ngFile)
	if buffer, err := c.StderrPipe(); err != nil {
		panic(err)
	} else {
		_ = c.Start()
		loadedByte, err := ioutil.ReadAll(buffer)
		if err != nil {
			panic(err)
		}
		_ = c.Wait()
		return loadedByte
	}
}

func (pe *PhpExecuter) SetOkFile(okFile string) {
	pe.okFile = okFile
}
func (pe *PhpExecuter) SetNgFile(ngFile string) {
	pe.ngFile = ngFile
}
func (pe *PhpExecuter) WriteToNg(input string) int {
	size, err := FileOpen(pe.ngFile, input)
	if err != nil {
		log.Fatal(err)
	}
	return size
}

//func (pe *PhpExecuter) WriteToOk(input string) (int, error) {
//	return FileOpen(pe.okFile, input)
//}

//type PhpExecuterInterface interface {
//	SetPhpExcutePath(string) bool
//	SetOkFile(string) bool
//	SetNgFile(string) bool
//	ExecutePhp() string
//	LoadBuffer() []byte
//	GetFatalError() []byte
//}

// PHPのfile関数と同様の処理をエミュレーション
func File(filePath string) ([]string, error) {
	var fileRows []string = make([]string, 0, 512)
	//fmt.Printf("len(fileRows): %v\n", len(fileRows))
	// 引数に渡されたファイルを読みこむ
	fp, err := os.Open(filePath)
	if err != nil {
		return []string{}, err
	}
	allBuffer, err := io.ReadAll(fp)
	// Handling error.
	if err != nil {
		return []string{}, err
	}

	var singleRow []byte
	var rowsNumber int = 0

	for _, value := range allBuffer {
		if string(value) == ("\n") {
			//fmt.Println(string(singleRow))
			fileRows = append(fileRows, string(singleRow))
			// 1行分をリセット
			singleRow = []byte{}
			rowsNumber++
			continue
		}
		singleRow = append(singleRow, value)
	}
	//fmt.Printf("fileRows: %v\n", fileRows)
	fileRows = fileRows[:rowsNumber]
	return fileRows, nil
}
