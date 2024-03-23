package wiredrawing

import (
	"os/exec"
)

func init() {
	// fmt.Println("wire-drawing パッケージ init関数")
}

func ValidateVNgFile(filePath string) int {
	command := exec.Command("php", filePath)
	// コマンドん実行結果をステータスコードで取得する
	var err = command.Run()
	if err != nil {
		//panic(err)
	}
	var exitCode int = command.ProcessState.ExitCode()
	return exitCode
}
