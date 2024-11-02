package wiredrawing

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/glebarez/go-sqlite"
	"github.com/hymkor/go-multiline-ny"
	"github.com/mattn/go-colorable"
	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/simplehistory"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"phpgo/config"
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

// StdInput ----------------------------------------
// 標準入力から入力された内容を文字列で返却する
// ----------------------------------------

func StdInput(prompt string, previousInput string) string {

	ctx := context.Background()
	//fmt.Println("C-m or Enter      : Insert a linefeed")
	//fmt.Println("C-p or UP         : Move to the previous line.")
	//fmt.Println("C-n or DOWN       : Move to the next line")
	//fmt.Println("C-j               : Submit")
	//fmt.Println("C-c               : Abort.")
	//fmt.Println("C-D with no chars : Quit.")
	//fmt.Println("C-UP   or M-P     : Move to the previous history entry")
	//fmt.Println("C-DOWN or M-N     : Move to the next history entry")

	var ed multiline.Editor
	var r readline.Coloring
	ed.SetColoring(r)
	ed.SetPrompt(func(w io.Writer, lnum int) (int, error) {
		return fmt.Fprintf(w, "[%d] ", lnum+1)
	})
	ed.SubmitOnEnterWhen(func(lines []string, index int) bool {
		if len(lines) == 1 {
			var f string = lines[0]
			if (f == "exit") || (f == "cat") || f == "yes" || f == "rollback" {
				return true
			}
		}
		for number, value := range lines {
			if (number == 0) && (value == "exit") {
				return true
			}
		}
		if index >= 1 {
			if (lines[index-1] == "") && (lines[index] == "") {
				return true
			}
		}
		return false

		//fmt.Printf("lines: %v\n", lines)
		//fmt.Printf("int => %v", index)
		//return strings.HasSuffix(strings.TrimSpace(lines[len(lines)-1]), ";")
	})
	// To enable escape sequence on Windows.
	// (On other operating systems, it can be ommited)
	ed.SetWriter(colorable.NewColorableStdout())

	history := simplehistory.New()
	ed.SetHistory(history)
	ed.SetHistoryCycling(true)

	for {
		lines, err := ed.Read(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return strings.Join(lines, "")
		}
		L := strings.Join(lines, "\n")
		//fmt.Println("-----")
		//fmt.Println(L)
		//fmt.Println("-----")
		history.Add(L)
		var stripLines []string
		for _, value := range lines {
			if strings.TrimSpace(value) != "" {
				stripLines = append(stripLines, value)
			}
		}
		return strings.Join(stripLines, "\n")
	}

	// 入力モードの選択用入力
	var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
	var result bool = scanner.Scan()
	if result != true {
		return ""
	}
	var which string = scanner.Text()
	if which != ">>>" {
		return which
	}
	fmt.Print("\033[33m")
	{
		var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
		var readString []string
		var s = 0
		for {
			//if s > 0 {
			//	fmt.Printf("%s%s", " ... ", " ... ")
			//} else {
			//	fmt.Printf("%s%s", " ... ", " >>> ")
			//}
			scanner.Scan()
			value := scanner.Text()

			//if value == "rollback" {
			//	if len(readString) > 0 {
			//		var lastString string = readString[len(readString)-1]
			//		readString = readString[0 : len(readString)-1]
			//		fmt.Print("\v" + colorWrapping("31", lastString) + "\n")
			//		continue
			//	}
			//}
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
	}
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
	okFileFp           *os.File
	ngFile             string
	ngFileFp           *os.File
	previousLine       int
	connection         net.Conn
	// 許容可能なエラーメッセージかどうか
	isAllowable bool
	// アプリケーション起動時からの全エラーメッセージを保持する
	wholeErrors  []string
	db           *sql.DB
	DatabasePath string
}

func (p *PhpExecuter) nextId() int {
	// 一時的にローカル変数に
	var db *sql.DB = p.db
	var nextId int
	tx, _ := db.Begin()
	rows, _ := tx.Query("select max(id) from phptext limit 1")
	for rows.Next() {
		_ = rows.Scan(&nextId)
		nextId++
	}
	// 意味はないけどcommit
	_ = tx.Commit()
	return nextId
}
func (p *PhpExecuter) currentId() int {
	var db *sql.DB = p.db
	var currentId int = 0
	tx, _ := db.Begin()
	rows, _ := tx.Query("select max(id) from phptext")
	for rows.Next() {
		err := rows.Scan(&currentId)
		if err != nil {
			log.Fatal(err)
		}
	}
	err := tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	return currentId
}
func (p *PhpExecuter) InitDB() *sql.DB {
	// sqliteの初期化R
	path, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	var dbPath string = path + "/" + ".php.db"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) != true {
		_ = os.Remove(dbPath)
	}
	db, err := sql.Open("sqlite", dbPath)
	p.DatabasePath = dbPath
	p.db = db
	if err != nil {
		panic(err)
	}

	createSql := `
	create table phptext (
	    id integer not null,
	    text text not null ,
	    is_production int
	)`
	_, err = db.Exec(createSql)
	if err != nil {
		panic(err)
	}
	// 実行結果を格納するテーブルを作成
	var sql string = `
		create table results (
		    id integer not null primary key autoincrement,
		    result text not null,
		    phptext_id integer not null
		)`
	// Make a table to save executed result.
	_, err = db.Exec(sql)
	if err != nil {
		log.Fatal(err)
	}
	nextId := p.nextId()
	tx, _ := db.Begin()
	st, _ := db.Prepare("insert into phptext(id, text, is_production) values (?, ?, ?)")
	st.Exec(nextId, "<?php", 1)
	tx.Commit()
	return db
}

// WholeErrors ----------------------------------------
func (pe *PhpExecuter) WholeErrors() []string {
	return pe.wholeErrors
}

// ResetWholeErrors ----------------------------------------
// 溜まったエラーメッセージをリセットする
func (pe *PhpExecuter) ResetWholeErrors() {
	pe.wholeErrors = []string{}
}

func (pe *PhpExecuter) Cat() []map[string]interface{} {
	db := pe.db
	query, err := db.Query("select id, text from phptext order by id asc")
	if err != nil {
		log.Fatal(err)
	}
	var logs []map[string]interface{}
	for query.Next() {
		var id int
		var text string
		err := query.Scan(&id, &text)
		if err != nil {
			log.Fatal(err)
		}
		var tempMap map[string]interface{} = map[string]interface{}{
			"id":   id,
			"text": text,
		}
		logs = append(logs, tempMap)
	}
	return logs
}

// SetPreviousList ----------------------------------------
// 前回のセーブポイントを変更する
func (pe *PhpExecuter) SetPreviousList(number int) int {
	var currenetLine int = pe.previousLine
	pe.previousLine = number
	return currenetLine
}
func (pe *PhpExecuter) GetPreviousList() int {
	var currenetLine int = pe.previousLine
	return currenetLine
}
func (pe *PhpExecuter) SetPhpExcutePath(phpPath string) {
	if phpPath == "" {
		pe.PhpPath = "php"
	}
	pe.PhpPath = phpPath
}

func (pe *PhpExecuter) Execute(showBuffer bool) (int, error) {
	logs := pe.Cat()
	phpLogs := ""
	for index := range logs {
		phpLogs += logs[index]["text"].(string) + "\n"
	}
	fp, _ := os.OpenFile(pe.okFile, os.O_RDWR, 0777)
	fp.Truncate(0)
	fp.Seek(0, 0)
	fp.WriteString(phpLogs)
	var colorCode string = config.Blue
	//// 一旦okFileFpを閉じるff
	//err := pe.okFileFp.Close()
	//if err != nil {
	//	log.Fatal(err)
	//}
	// isValidate == true の場合はngFileを実行(事前実行)
	command := exec.Command(pe.PhpPath, fp.Name())

	buffer, err := command.StdoutPipe()
	if err != nil {
		return 0, err
	}
	err = command.Start()
	if err != nil {
		return 0, err
	}
	var currentLine int

	const ensureLength int = 4096

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

		//// 実行結果としてSqliteに保存する
		//pe.WriteResultToDB(string(readData))

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
	//// 再度新規pointerとしてokFileFpを開く
	//pe.okFileFp, err = os.OpenFile(pe.okFile, os.O_RDWR, 0777)
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
	//const ParseErrorString = `PHP[ ]+?Parse[ ]+?error:[ ]+?syntax[ ]+?error`
	const ParseErrorString = `PHP[ ]+?Parse[ ]+?error:[ ]+?`
	var parseErrorRegex = regexp.MustCompile(ParseErrorString)
	pe.IsPermissibleError = false

	//// 事前にPHPコマンド <php -l file-name>でシンタックスエラーのみを先にチェックする
	//syntax := exec.Command(pe.PhpPath, "-l", pe.ngFile)
	//syntexBuffer, err := syntax.StderrPipe()
	//if err != nil {
	//	panic(err)
	//}
	//_ = syntax.Start()
	//// シンタックスエラーがでている場合
	//loaded, err := io.ReadAll(syntexBuffer)
	//if err != nil {
	//	panic(err)
	//}
	//_ = syntax.Wait()
	//if syntax.ProcessState.ExitCode() != 0 {
	//	// シンタックスエラーがあった場合
	//	//fmt.Printf("Syntax Error: <<<%v>>>\n", string(loaded))
	//	pe.IsPermissibleError = true
	//	pe.ErrorBuffer = loaded
	//	pe.wholeErrors = append(pe.wholeErrors, string(loaded))
	//	return true, nil
	//}

	logs := pe.Cat()
	phpLogs := ""
	for index := range logs {
		phpLogs += logs[index]["text"].(string) + "\n"
	}
	fp, _ := os.OpenFile(pe.ngFile, os.O_RDWR, 0777)
	fp.Truncate(0)
	fp.Seek(0, 0)
	fp.WriteString(phpLogs)
	// 終了コードが不正な場合,FatalErrorを取得する
	c := exec.Command(pe.PhpPath, fp.Name())
	buffer, err := c.StderrPipe()
	if err != nil {
		fmt.Printf("err in DetectFatalError: %v\n", err)
	}
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
			pe.wholeErrors = append(pe.wholeErrors, string(loadedByte))
			return false, nil
			//return loadedByte, false, nil
		}
		// 終了コードが正常な場合,何もしない
		pe.ErrorBuffer = []byte{}
		return false, nil
	}
	// エラー内容がシンタックスエラーなら許容する
	if parseErrorRegex.MatchString(string(loadedByte)) {
		pe.IsPermissibleError = false
	}
	// シンタックスエラーのみ許容するが Fatal Error in PHP である
	pe.ErrorBuffer = loadedByte
	pe.wholeErrors = append(pe.wholeErrors, string(loadedByte))
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
	if pe.okFile == "" {
		pe.okFile = okFile
	}
	if pe.okFileFp == nil {
		//fp, err := os.OpenFile(pe.okFile, os.O_RDWR, 0777)
		//if err != nil {
		//	log.Fatal(err)
		//}
		//pe.okFileFp = fp
	} else {
		//pe.okFileFp.Close()
		//fp, err := os.OpenFile(pe.okFile, os.O_RDWR, 0777)
		//if err != nil {
		//	log.Fatal(err)
		//}
		//pe.okFileFp = fp
	}
}
func (pe *PhpExecuter) SetNgFile(ngFile string) {
	if pe.ngFile == "" {
		pe.ngFile = ngFile
	}
	if pe.ngFileFp == nil {
		//fp, err := os.OpenFile(pe.ngFile, os.O_RDWR, 0777)
		//if err != nil {
		//	log.Fatal(err)
		//}
		//pe.ngFileFp = fp
	} else {
		//pe.ngFileFp.Close()
		//fp, err := os.OpenFile(pe.ngFile, os.O_RDWR, 0777)
		//if err != nil {
		//	log.Fatal(err)
		//}
		//pe.ngFileFp = fp
	}
}

// PHPファイルの実行結果をSqliteに保存
func (p *PhpExecuter) WriteResultToDB(result string) bool {
	db := p.db
	// Start transaction.
	tx, _ := db.Begin()
	rows, err := tx.Query("select max(id) from phptext")
	if err != nil {
		log.Fatal(err)
	}
	if rows.Next() != true {
		log.Fatal("Failed fetching lastest primary key on phptext table.")
	}
	var lastestId int = 0
	if err := rows.Scan(&lastestId); err != nil {
		log.Fatal(err)
	}

	statement, _ := tx.Prepare("insert into results (result, phptext_id) values (?, ?)")
	if _, err := statement.Exec(result, lastestId); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	return true
}

// WriteToDB 指定したテキストをSqliteに書き込む
func (pe *PhpExecuter) WriteToDB(input string, isProduction int) int64 {
	// sqliteへ書き込む
	tx, _ := pe.db.Begin()
	st, err := tx.Prepare("insert into phptext(id, text, is_production) values (?, ?, ?)")
	if err != nil {
		panic(err)
	}
	if int(input[0]) == 27 {
		tx.Rollback()
		return 0
	}
	// 取得したnextID, 本文, 実行するタイミング
	cleansing := strings.TrimRight(input, "\r\n ")
	result, _ := st.Exec(pe.nextId(), cleansing, isProduction)
	latestId, err := result.LastInsertId()
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
	return latestId
}

func (pe *PhpExecuter) WriteToNg(input string) int64 {
	//var err error = nil
	//// ngFileのポインタを末尾に移動させる
	//_, _ = io.ReadAll(pe.ngFileFp)
	//size, err := pe.ngFileFp.WriteString(input)
	//if err != nil {
	//	log.Fatal(err)
	//}
	latestId := pe.WriteToDB(input, 0)
	return latestId
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
	defer (func() {
		_ = src.Close()
		_ = des.Close()
	})()
	_, err := io.Copy(des, src)
	if err != nil {
		panic(err)
	}
	return true
}

//func (pe *PhpExecuter) CopyFromNgToOk() (int, []byte) {
//	_, err := pe.ngFileFp.Seek(0, 0)
//	if err != nil {
//		log.Fatal(err)
//	}
//	allText, err := io.ReadAll(pe.ngFileFp)
//	if err != nil {
//		log.Fatal(err)
//	}
//	_ = pe.okFileFp.Truncate(0)
//	size, err := pe.okFileFp.Write(allText)
//	if err != nil {
//		log.Fatal(err)
//	}
//	return size, allText
//}

// Rollback ----------------------------------------
// OkFileの中身をNgFileまるっとコピーする
func (pe *PhpExecuter) Rollback() int {
	var size int = 0
	// ロールバック処理
	// ファイルの内容を全て削除する
	//_ = pe.ngFileFp.Truncate(0)
	//_, err := pe.ngFileFp.Seek(0, 0)
	//if err != nil {
	//	log.Fatalf("前実行用ファイルのポインタを先頭に移動できませんでした: [%v]", err)
	//}
	//// OkFileのファイルポインタを先頭に移す
	//_, err = pe.okFileFp.Seek(0, 0)
	//if err != nil {
	//	log.Fatalf("後実行用ファイルのポインタを先頭に移動できませんでした: [%v]", err)
	//}
	//all, _ := io.ReadAll(pe.okFileFp)
	//size, _ := pe.ngFileFp.Write(all)

	var db *sql.DB = pe.db
	var err error = nil
	tx, _ := db.Begin()
	statment, _ := tx.Prepare("delete from phptext where id = ?")
	_, _ = statment.Exec(pe.currentId())
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	return size
}
func (pe *PhpExecuter) Clear() bool {
	//_ = pe.ngFileFp.Truncate(0)
	//_ = pe.okFileFp.Truncate(0)
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
