package wiredrawing

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/glebarez/go-sqlite"
	"github.com/hymkor/go-multiline-ny"
	"github.com/mattn/go-colorable"
	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/keys"
	"github.com/nyaosorg/go-readline-ny/simplehistory"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"phpgo/config"
	. "phpgo/errorhandler"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
	"unsafe"
)

const (
	AltEnter = "\x1B\r"
)

type PHPSource struct {
	text     string
	sourceId int
}
type MyBinder struct {
	ctrlCNum int
	prompt   int
}

func (my *MyBinder) SetPrompt(prompt int) {
	my.prompt = prompt
}
func (my *MyBinder) Prompt() int {
	return my.prompt
}
func (my *MyBinder) Call(ctx context.Context, b *readline.Buffer) readline.Result {
	if my.prompt > 0 {
		my.prompt += 1
		return readline.ENTER
	} else {
		my.prompt = 1
		fmt.Println(config.ColorWrapping(config.Red, "終了する場合はもう一度 ctrl+C を押して下さい"))
		return readline.ENTER
	}
}
func (my *MyBinder) String() string {
	return "MyBinder"

}

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

var myBinder *MyBinder = &MyBinder{
	ctrlCNum: 0,
	prompt:   0,
}
var ed multiline.Editor

// StdInput ----------------------------------------
// 標準入力から入力された内容を文字列で返却する
// ----------------------------------------

func StdInput(prompt string, previousInput string) (string, int) {

	ctx := context.Background()
	type ac = readline.AnonymousCommand

	Catch(ed.BindKey(keys.Delete, ac(ed.CmdBackwardDeleteChar)))
	Catch(ed.BindKey(keys.Backspace, ac(ed.CmdBackwardDeleteChar)))
	Catch(ed.BindKey(keys.Backspace, ac(ed.CmdBackwardDeleteChar)))
	Catch(ed.BindKey(AltEnter, readline.AnonymousCommand(ed.Submit)))
	Catch(ed.BindKey(keys.CtrlBackslash, readline.AnonymousCommand(ed.Submit)))
	Catch(ed.BindKey(keys.CtrlZ, readline.AnonymousCommand(ed.Submit)))
	//ed.BindKey(keys.CtrlC, myBinder)
	if len(previousInput) > 0 {
		if previousInput != "exit" && previousInput != "rollback" && previousInput != "cat" && previousInput != "save" {
			splitPreviousInput := strings.Split(previousInput, "\n")
			ed.SetDefault(splitPreviousInput)
		}
	} else {
		ed.SetDefault(nil)
	}
	ed.SetPrompt(func(w io.Writer, lnum int) (int, error) {
		return fmt.Fprintf(w, "\033[0m%d:>>> ", lnum+1)
	})
	ed.SubmitOnEnterWhen(func(lines []string, index int) bool {
		fmt.Print("\033[33m")
		// strip input text.
		var replaceLines []string
		for _, v := range lines {
			stripV := strings.TrimSpace(v)
			if len(stripV) == 0 {
				continue
			}
			replaceLines = append(replaceLines, stripV)
		}
		if len(replaceLines) == 0 {
			return false
		}
		// 最後の行が<\c>で終わっている場合はtrueを返却する
		if (len(replaceLines) > 0) && replaceLines[len(replaceLines)-1] == "\\c" {
			return true
		}

		if len(replaceLines) == 1 {
			var f string = replaceLines[0]
			if (f == "exit") || (f == "cat") || f == "yes" || f == "rollback" || f == "save" {
				return true
			}
		}
		for number, value := range lines {
			if (number == 0) && (value == "exit") {
				return true
			}
		}
		connected := strings.Join(replaceLines, "")
		// 入力内容の末尾が_(アンダースコア)で完了している場合はtrueを返却
		if strings.HasSuffix(connected, "_") {
			return true
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
		fmt.Print("\033[0m")
		lines, err := ed.Read(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return strings.Join(lines, ""), myBinder.Prompt()
		}
		if len(lines) == 0 {
			return "", myBinder.Prompt()
		}

		var fixed []string
		for _, value := range lines {
			if len(value) > 0 {
				fixed = append(fixed, value)
			}
		}
		if len(fixed) == 0 {
			return "", myBinder.Prompt()
		}
		if fixed[len(fixed)-1] == "\\c" {
			fixed = fixed[:len(fixed)-1]
		}
		L := strings.Join(fixed, "\n")
		//fmt.Println("-----")
		//fmt.Println(L)
		//fmt.Println("-----")
		history.Add(L)
		var stripLines []string
		for _, value := range fixed {
			if strings.TrimSpace(value) != "" {
				stripLines = append(stripLines, value)
			}
		}
		result := strings.Join(stripLines, "\n")
		if strings.HasSuffix(result, "_") {
			result = result[:len(result)-1]
		}
		return result, myBinder.Prompt()
	}
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
	hashKey      string
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

func (p *PhpExecuter) makeHashKey() string {
	var today time.Time = time.Now()
	// 保存用のハッシュを作詞絵
	rand.Seed(today.Unix())
	var source []byte = make([]byte, 256)
	for i := 0; i < 256; i++ {
		source[i] = byte(rand.Intn(255))
	}
	first := sha512.New()
	first.Write(source)
	hashKey := fmt.Sprintf("%x", first.Sum(nil))
	return hashKey
}
func (p *PhpExecuter) InitDB() *sql.DB {

	p.hashKey = p.makeHashKey()
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
	Catch(err)

	backUpTable := `
		create table backup_phptext(
		    id integer not null primary key  autoincrement,
		    phptext_id integer not null,
		    text text not null,
		    created_at datetime not null DEFAULT (DATETIME('now', 'localtime')),
		    key text not null
		)
	`
	_, err = db.Exec(backUpTable)
	Catch(err)

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
func (p *PhpExecuter) WholeErrors() []string {
	return p.wholeErrors
}

// ResetWholeErrors ----------------------------------------
// 溜まったエラーメッセージをリセットする
func (p *PhpExecuter) ResetWholeErrors() {
	p.wholeErrors = []string{}
}

func (p *PhpExecuter) Cat() []map[string]interface{} {
	db := p.db
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
func (p *PhpExecuter) SetPreviousList(number int) int {
	var currenetLine int = p.previousLine
	p.previousLine = number
	return currenetLine
}
func (p *PhpExecuter) GetPreviousList() int {
	var currenetLine int = p.previousLine
	return currenetLine
}
func (p *PhpExecuter) SetPhpExcutePath(phpPath string) {
	if phpPath == "" {
		p.PhpPath = "php"
	}
	p.PhpPath = phpPath
}

func (p *PhpExecuter) Execute(showBuffer bool) (int, error) {
	logs := p.Cat()
	phpLogs := ""
	for index := range logs {
		phpLogs += logs[index]["text"].(string) + "\n"
	}
	fp, _ := os.OpenFile(p.okFile, os.O_RDWR, 0777)
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
	command := exec.Command(p.PhpPath, fp.Name())

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
	fmt.Print("\033[0m")
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
		if (currentLine + n) >= p.previousLine {
			if from < p.previousLine && p.previousLine <= to {
				diff := p.previousLine - currentLine
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
	p.previousLine = currentLine
	_ = command.Wait()
	// 使用したメモリを開放してみる
	runtime.GC()
	debug.FreeOSMemory()
	// コンソールのカラーをもとにもどす
	_, _ = os.Stdout.WriteString("\033[0m")
	//debug.FreeOSMemory()
	p.ErrorBuffer = []byte{}
	//// 再度新規pointerとしてokFileFpを開く
	//pe.okFileFp, err = os.OpenFile(pe.okFile, os.O_RDWR, 0777)
	return outputSize, nil
}

// DetectFatalError ----------------------------------------
// 事前にPHPの実行結果がエラーであるかどうかを判定する
func (p *PhpExecuter) DetectFatalError() (bool, error) {

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
	p.IsPermissibleError = false

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

	logs := p.Cat()
	phpLogs := ""
	for index := range logs {
		phpLogs += logs[index]["text"].(string) + "\n"
	}
	fp, _ := os.OpenFile(p.ngFile, os.O_RDWR, 0777)
	fp.Truncate(0)
	fp.Seek(0, 0)
	fp.WriteString(phpLogs)
	// 終了コードが不正な場合,FatalErrorを取得する
	c := exec.Command(p.PhpPath, fp.Name())
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
			p.ErrorBuffer = loadedByte
			p.wholeErrors = append(p.wholeErrors, string(loadedByte))
			return false, nil
			//return loadedByte, false, nil
		}
		// 終了コードが正常な場合,何もしない
		p.ErrorBuffer = []byte{}
		return false, nil
	}
	// エラー内容がシンタックスエラーなら許容する
	if parseErrorRegex.MatchString(string(loadedByte)) {
		p.IsPermissibleError = false
	}
	// シンタックスエラーのみ許容するが Fatal Error in PHP である
	p.ErrorBuffer = loadedByte
	p.wholeErrors = append(p.wholeErrors, string(loadedByte))
	return true, nil
}

func (p *PhpExecuter) DetectErrorExceptFatalError() ([]byte, error) {
	c := exec.Command(p.PhpPath, p.ngFile)
	buffer, err := c.StderrPipe()
	if err != nil {
		return []byte{}, err
	}
	_ = c.Start()
	loadedByte, err := ioutil.ReadAll(buffer)
	_ = c.Wait()
	return loadedByte, nil
}

func (p *PhpExecuter) GetFatalError() []byte {
	c := exec.Command(p.PhpPath, p.ngFile)
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

func (p *PhpExecuter) SetOkFile(okFile string) {
	if p.okFile == "" {
		p.okFile = okFile
	}
	if p.okFileFp == nil {
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
func (p *PhpExecuter) SetNgFile(ngFile string) {
	if p.ngFile == "" {
		p.ngFile = ngFile
	}
	if p.ngFileFp == nil {
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
func (p *PhpExecuter) WriteToDB(input string, isProduction int) int64 {
	// sqliteへ書き込む
	tx, _ := p.db.Begin()
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
	result, _ := st.Exec(p.nextId(), cleansing, isProduction)
	latestId, err := result.LastInsertId()
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
	return latestId
}

func (p *PhpExecuter) WriteToNg(input string) int64 {
	//var err error = nil
	//// ngFileのポインタを末尾に移動させる
	//_, _ = io.ReadAll(pe.ngFileFp)
	//size, err := pe.ngFileFp.WriteString(input)
	//if err != nil {
	//	log.Fatal(err)
	//}
	latestId := p.WriteToDB(input, 0)
	return latestId
}

func (p *PhpExecuter) Save(saveFileName string) bool {
	current := make([]PHPSource, 0, 64)
	sql := `
		select id, text from phptext order by id asc
		`
	tx, err := p.db.Begin()
	Catch(err)
	rows, err := tx.Query(sql)
	Catch(err)
	var sourceText string = ""
	var sourceId int = 0
	currentConnected := ""
	for rows.Next() {
		Catch(rows.Scan(&sourceId, &sourceText))
		fmt.Printf("sourceText => %v", sourceText)
		currentConnected += sourceText + "\n"
		current = append(current, PHPSource{text: sourceText, sourceId: sourceId})
	}

	fmt.Printf("current => %v", current)
	// 保存のたびにハッシュキーを計算
	p.hashKey = p.makeHashKey()
	for _, row := range current {
		sql := "insert into backup_phptext (phptext_id, text, key) values(? ,?, ?)"
		statement, err := tx.Prepare(sql)
		Catch(err)
		_, err = statement.Exec(row.sourceId, row.text, p.hashKey)
		Catch(err)
	}
	Catch(tx.Commit())
	// バックアップ用ファイルを作成する
	wd, _ := os.Getwd()
	if saveFileName == "" {
		saveFileName = "save.php"
	} else {
		saveFileName = wd + "/" + saveFileName
	}
	if _, err := os.Stat(saveFileName); os.IsNotExist(err) != true {
		_ = os.Remove(saveFileName)
	}
	des, err := os.OpenFile(saveFileName, os.O_CREATE|os.O_RDWR, 0777)
	Catch(err)
	_, err = des.WriteString(currentConnected)
	Catch(err)
	defer (func() {
		_ = des.Close()
	})()
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
func (p *PhpExecuter) Rollback() int {
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

	var db *sql.DB = p.db
	var err error = nil
	tx, _ := db.Begin()
	statment, _ := tx.Prepare("delete from phptext where id = ?")
	_, _ = statment.Exec(p.currentId())
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	return size
}
func (p *PhpExecuter) Clear() bool {
	//_ = pe.ngFileFp.Truncate(0)
	//_ = pe.okFileFp.Truncate(0)
	return true
}
