package wiredrawing

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
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
	connection         net.Conn
	// 許容可能なエラーメッセージかどうか
	isAllowable bool
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

func (pe *PhpExecuter) Execute(showBuffer bool) (int, error) {

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

	const ensureLength int = 256

	currentLine = 0
	var outputSize int = 0
	// whenError == true の場合バッファ内容を返却してやる
	//var bufferWhenError string
	_, _ = os.Stdout.WriteString("\033[" + colorCode + "m")
	//t := time.Now()
	//formatted := t.Format(time.RFC3339)
	//_, _ = os.Stdout.WriteString(formatted + " ")
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
				outputSize += len(tempSlice)
				if showBuffer == true {
					_, err = os.Stdout.WriteString(*(*string)(unsafe.Pointer(&tempSlice)))
					if err != nil {
						log.Fatal(err)
					}
				}
			} else {
				// 出力内容の表示フラグがtrueの場合のみ
				outputSize += len(readData)
				if showBuffer == true {
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
	pe.ErrorBuffer = []byte{}
	return outputSize, nil
}

// DetectFatalError ----------------------------------------
// 事前にPHPの実行結果がエラーであるかどうかを判定する
func (pe *PhpExecuter) DetectFatalError() (bool, error) {

	defer func() {
		if err := recover(); err != nil {
			if err, ok := err.(error); ok {
				fmt.Println(err)
				return
				//return []byte{}, err
			}
		}
		return
	}()
	//panic(errors.New("意図しないエラー"))
	// PHPのエラーメッセージの正規表現を事前コンパイルする
	const ParseErrorString = `PHP[ ]+?Parse[ ]+?error:[ ]+?syntax[ ]+?error`
	var parseErrorRegex = regexp.MustCompile(ParseErrorString)
	pe.IsPermissibleError = false

	// 事前にPHPコマンド <php -l file-name>でシンタックスエラーのみを先にチェックする
	syntax := exec.Command(pe.PhpPath, "-l", pe.ngFile)
	syntexBuffer, err := syntax.StderrPipe()
	if err != nil {
		panic(err)
	}
	_ = syntax.Start()
	// シンタックスエラーがでている場合
	loaded, err := io.ReadAll(syntexBuffer)
	if err != nil {
		panic(err)
	}
	_ = syntax.Wait()
	if syntax.ProcessState.ExitCode() != 0 {
		// シンタックスエラーがあった場合
		//fmt.Printf("Syntax Error: <<<%v>>>\n", string(loaded))
		pe.IsPermissibleError = true
		pe.ErrorBuffer = loaded
		return true, nil
	}

	// 終了コードが不正な場合,FatalErrorを取得する
	c := exec.Command(pe.PhpPath, pe.ngFile)
	buffer, _ := c.StderrPipe()
	// 戻り値自体がインターフェースである以上,*os.File型へは代入できない
	// そのためどうしても具象型にしたい場合は型アサーションを使う
	buffer, ok := buffer.(*os.File)
	if ok != true {
		panic(errors.New("failed to convert io.Reader to *os.File"))
	}
	_ = c.Start()
	loadedByte, err := io.ReadAll(buffer)
	if err != nil {
		// 実行結果の出力, PHPのFatal Errorかどうか, Goのエラー
		return false, err
		//return []byte{}, true, err
	}
	_ = c.Wait()

	//fmt.Printf("ExitCode: %v\n", c.ProcessState.ExitCode())
	if c.ProcessState.ExitCode() == 0 {
		if len(loadedByte) > 0 {
			// Fatal Error in PHP ではない
			// また標準エラー出力はオブジェクトから取得する
			pe.ErrorBuffer = loadedByte
			return false, nil
			//return loadedByte, false, nil
		}
		// 終了コードが正常な場合,何もしない
		pe.ErrorBuffer = []byte{}
		return false, nil
	}
	// エラー内容がシンタックスエラーなら許容する
	if parseErrorRegex.MatchString(string(loadedByte)) {
		pe.IsPermissibleError = true
	}
	// シンタックスエラーのみ許容するが Fatal Error in PHP である
	pe.ErrorBuffer = loadedByte
	return true, nil
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

func (pe *PhpExecuter) Save(saveFileName string) bool {
	// バックアップ用ファイルを作成する
	wd, _ := os.Getwd()
	if saveFileName == "" {
		saveFileName = "save.php"
	} else {
		saveFileName = wd + "/" + saveFileName
	}
	des, _ := os.OpenFile(saveFileName, os.O_CREATE, 0777)
	src, _ := os.OpenFile(pe.ngFile, os.O_RDONLY, 0777)
	_, err := io.Copy(des, src)
	if err != nil {
		panic(err)
	}
	_ = src.Close()
	_ = des.Close()
	return true
}

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

//func (pe *PhpExecuter) passTCPServer(s string) {
//
//	if pe.connection == nil {
//		addr := net.TCPAddr{
//			IP:   net.ParseIP("127.0.0.1"),
//			Port: 8888,
//		}
//		connect, err := net.DialTCP("tcp", nil, &addr)
//		if err != nil {
//			panic(err)
//		}
//		pe.connection = connect
//	}
//	_, _ = pe.connection.Write([]byte(s))
//	//for {
//	//	select {
//	//	case s := <-stringChannel:
//	//		_, _ = pe.connection.Write([]byte(s))
//	//	}
//	//}
//}
