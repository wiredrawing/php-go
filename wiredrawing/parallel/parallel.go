package parallel

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

// InterruptProcess
// 割り込み対処を実行するGoルーチン
// コンソール上でのinterruptイベントを監視
// --------------------------------------------
func InterruptProcess(exit chan int, observer chan os.Signal) {
	var s os.Signal
	for {
		// 変数 observerから中身を取り出すまで ここで処理がとまる
		s, _ = <-observer // 待機
		if s == syscall.SIGHUP {
			fmt.Print("[syscall.SIGHUP].\r\n")
			// 割り込みを無視
			exit <- 1
		} else if s == syscall.SIGTERM {
			fmt.Print("[syscall.SIGTERM].\r\n")
			exit <- 2
		} else if s == os.Kill {
			fmt.Print("[os.Kill].\r\n")
			// 割り込みを無視
			exit <- 3
		} else if s == os.Interrupt {
			// 割り込みを無視
			exit <- 4
		} else if s == syscall.Signal(0x14) {
			if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
				fmt.Print("[syscall.SIGTSTP].\r\n")
			}
			// 割り込みを無視
			exit <- 5
		} else if s == syscall.SIGQUIT {
			fmt.Print("[syscall.SIGQUIT].\r\n")
			exit <- 6
		} else {
			// 未定義の割り込み処理
			fmt.Print("[Unknown syscall].\r\n")
			exit <- -1
		}
		//// 変数 observerから中身を取り出すまで ここで処理がとまる
		//s, _ = <-observer // 待機
	}
}
