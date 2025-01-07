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
	"log"
	"math/rand"
	"os"
	"os/exec"
	"phpgo/config"
	. "phpgo/errorhandler"
	"regexp"
	//"runtime"
	//"runtime/debug"
	"strings"
	"time"
)

type PHPSource struct {
	text     string
	sourceId int
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

var ed multiline.Editor

var worsToExis []string = []string{
	"exit",
	"cat",
	"yes",
	"rollback",
	"save",
	"clear",
	"\\c",
}

// StdInput ----------------------------------------
// 標準入力から入力された内容を文字列で返却する
// ----------------------------------------
func StdInput(prompt string, previousInput string, p *PHPExecuter) (string, int) {
	ctx := context.Background()
	type ac = readline.AnonymousCommand

	Catch(ed.BindKey(keys.Delete, ac(ed.CmdBackwardDeleteChar)))
	Catch(ed.BindKey(keys.Backspace, ac(ed.CmdBackwardDeleteChar)))
	Catch(ed.BindKey(keys.Backspace, ac(ed.CmdBackwardDeleteChar)))
	Catch(ed.BindKey(keys.CtrlBackslash, readline.AnonymousCommand(ed.Submit)))
	Catch(ed.BindKey(keys.CtrlZ, readline.AnonymousCommand(ed.Submit)))
	Catch(ed.BindKey(keys.CtrlUnderbar, readline.AnonymousCommand(ed.Submit)))
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
		// 最後の行が指定された値で終わっている場合はtrueを返却する
		var f string = replaceLines[len(replaceLines)-1]
		if InArray(f, worsToExis) {
			return true
		}
		connected := strings.Join(replaceLines, "")
		// 入力内容の末尾が_(アンダースコア)で完了している場合はtrueを返却
		if strings.HasSuffix(connected, "_") {
			return true
		}
		return false
	})
	// To enable escape sequence on Windows.
	// (On other operating systems, it can be ommited)
	ed.SetWriter(colorable.NewColorableStdout())

	history := simplehistory.New()
	ed.SetHistory(history)
	ed.SetHistoryCycling(true)

	for {
		fmt.Print(`[0m`)
		lines, err := ed.Read(ctx)
		if err != nil {
			Catch(fmt.Fprintln(os.Stderr, err.Error()))
			return strings.Join(lines, ""), 0
		}
		if len(lines) == 0 {
			return "", 0
		}

		var fixed []string
		for _, value := range lines {
			if len(value) > 0 {
				fixed = append(fixed, value)
			}
		}
		if len(fixed) == 0 {
			return "", 0
		}
		if fixed[len(fixed)-1] == "\\c" {
			fixed = fixed[:len(fixed)-1]
		}
		L := strings.Join(fixed, "\n")
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
		// 第二引数は<0>
		if len(result) > 0 {
			p.WriteToDB(result, 0)
		}
		return result, 0
	}
}

type PHPExecuter struct {
	PhpPath            string
	IsPermissibleError bool
	ErrorBuffer        []byte
	ngFile             string
	ngFileFp           *os.File
	previousLine       int
	// アプリケーション起動時からの全エラーメッセージを保持する
	wholeErrors []string
	db          *sql.DB
	hashKey     string
}

func (p *PHPExecuter) currentId() int {
	var db *sql.DB = p.db
	var currentId int = 0
	tx, _ := db.Begin()
	rows, _ := tx.Query("select max(id) from phptext where is_production = 1")
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

func (p *PHPExecuter) makeHashKey() string {
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
func (p *PHPExecuter) InitDB() *sql.DB {

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
	p.db = db
	if err != nil {
		panic(err)
	}

	createSql := `
	create table phptext (
	    id integer not null primary key autoincrement ,
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
	tx, _ := db.Begin()
	st, _ := db.Prepare("insert into phptext(text, is_production) values (?, ?)")
	st.Exec("<?php", 1)
	tx.Commit()
	return db
}

// WholeErrors ----------------------------------------
func (p *PHPExecuter) WholeErrors() []string {
	return p.wholeErrors
}

// ResetWholeErrors ----------------------------------------
// 溜まったエラーメッセージをリセットする
func (p *PHPExecuter) ResetWholeErrors() {
	p.wholeErrors = []string{}
}

func (p *PHPExecuter) Cat(isProd int) []map[string]interface{} {
	db := p.db
	statement, err := db.Prepare("select id, text from phptext where is_production = ? order by id asc")
	if err != nil {
		log.Fatal(err)
	}
	rows, _ := statement.Query(isProd)
	var logs []map[string]interface{}
	for rows.Next() {
		var id int
		var text string
		err := rows.Scan(&id, &text)
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
func (p *PHPExecuter) SetPreviousList(number int) int {
	var currenetLine int = p.previousLine
	p.previousLine = number
	return currenetLine
}
func (p *PHPExecuter) GetPreviousList() int {
	var currenetLine int = p.previousLine
	return currenetLine
}
func (p *PHPExecuter) SetPhpExcutePath(phpPath string) {
	if phpPath == "" {
		p.PhpPath = "php"
	}
	p.PhpPath = phpPath
}

func (p *PHPExecuter) Execute(showBuffer bool, isProd int) (int, error) {
	logs := p.Cat(isProd)
	phpLogs := ""
	for index := range logs {
		phpLogs += logs[index]["text"].(string) + "\n"
	}
	fp, _ := os.OpenFile(p.ngFile, os.O_RDWR, 0777)
	Catch(fp.Truncate(0))
	Catch(fp.Seek(0, 0))
	Catch(fp.WriteString(phpLogs))
	var colorCode string = config.Blue

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
	fmt.Print("\033[0m")
	_, _ = os.Stdout.WriteString("\033[" + colorCode + "m")
	for {
		readData := make([]byte, ensureLength)
		n, err := buffer.Read(readData)
		if (err != nil) && (err != io.EOF) {
			fmt.Printf(err.Error())
			readData = nil
			break
		}
		if n == 0 {
			readData = nil
			break
		}
		// 正味のバッファを取り出す
		readData = readData[:n]

		from := currentLine
		to := currentLine + n
		if (currentLine + n) >= p.previousLine {
			if from < p.previousLine && p.previousLine <= to {
				diff := p.previousLine - currentLine
				tempSlice := readData[diff:]
				// 出力内容の表示フラグがtrueの場合のみ
				outputSize += len(tempSlice)
				if showBuffer == true {
					Catch(fmt.Fprintf(os.Stdout, string(tempSlice)))
				}
			} else {
				// 出力内容の表示フラグがtrueの場合のみ
				outputSize += len(readData)
				if showBuffer == true {
					Catch(fmt.Fprintf(os.Stdout, string(readData)))
				}
			}
		}
		currentLine += n
		readData = nil
	}
	p.previousLine = currentLine
	_ = command.Wait()
	// 使用したメモリを開放してみる
	//runtime.GC()
	//debug.FreeOSMemory()
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
func (p *PHPExecuter) DetectFatalError(isProd int) (bool, error) {

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

	logs := p.Cat(isProd)
	phpLogs := ""
	for index := range logs {
		phpLogs += logs[index]["text"].(string) + "\n"
	}
	fp, _ := os.OpenFile(p.ngFile, os.O_RDWR, 0777)
	Catch(fp.Truncate(0))
	Catch(fp.Seek(0, 0))
	Catch(fp.WriteString(phpLogs))
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

func (p *PHPExecuter) DetectErrorExceptFatalError() ([]byte, error) {
	c := exec.Command(p.PhpPath, p.ngFile)
	buffer, err := c.StderrPipe()
	if err != nil {
		return []byte{}, err
	}
	_ = c.Start()
	loadedByte, err := io.ReadAll(buffer)
	//loadedByte, err := ioutil.ReadAll(buffer)
	_ = c.Wait()
	return loadedByte, nil
}

func (p *PHPExecuter) SetNgFile(ngFile string) {
	if p.ngFile == "" {
		p.ngFile = ngFile
	}
}

// WriteToDB 指定したテキストをSqliteに書き込む
func (p *PHPExecuter) WriteToDB(input string, isProduction int) int64 {
	// sqliteへ書き込む
	tx, _ := p.db.Begin()
	st, err := tx.Prepare("insert into phptext(text, is_production) values (?, ?)")
	if err != nil {
		panic(err)
	}
	if int(input[0]) == 27 {
		tx.Rollback()
		return 0
	}
	// 取得したnextID, 本文, 実行するタイミング
	cleansing := strings.TrimRight(input, "\r\n ")
	result, _ := st.Exec(cleansing, isProduction)
	latestId, err := result.LastInsertId()
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
	return latestId
}

func (p *PHPExecuter) WriteToNg(input string, isProd int) int64 {
	latestId := p.WriteToDB(input, isProd)
	return latestId
}

func (p *PHPExecuter) Save(saveFileName string) bool {
	current := make([]PHPSource, 0, 64)
	sql := `
		select id, text from phptext where is_production = 1 order  by id asc
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
		currentConnected += sourceText + "\n"
		current = append(current, PHPSource{text: sourceText, sourceId: sourceId})
	}

	//fmt.Printf("current => %v", current)
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

// Rollback ----------------------------------------
// OkFileの中身をNgFileまるっとコピーする
func (p *PHPExecuter) Rollback() bool {
	var db *sql.DB = p.db
	var err error = nil
	tx, _ := db.Begin()
	rows, _ := tx.Query("select count(id) from phptext where is_production = 1")

	var surplus int
	for rows.Next() {
		Catch(rows.Scan(&surplus))
	}
	if surplus > 1 {
		statment, _ := tx.Prepare("delete from phptext where id = ?")
		_, _ = statment.Exec(p.currentId())
		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
		return true
	}
	_ = tx.Commit()
	return false
}
func (p *PHPExecuter) Clear() bool {
	var db = p.db

	tx, _ := db.Begin()
	result, err := tx.Exec("delete from phptext")
	if err != nil {
		return false
	}
	affectedNum, _ := result.RowsAffected()
	fmt.Printf("Deleted %v records.", affectedNum)
	tx.Commit()
	return true
}
