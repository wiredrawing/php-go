package parallel

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

// --------------------------------------------
// 割り込み対処を実行するGoルーチン
// コンソール上でのinterruptイベントを監視
// --------------------------------------------
func InterruptProcess(exit chan int, observer chan os.Signal) {
	echo := fmt.Print
	var s os.Signal
	for {
		s, _ = <-observer
		if s == syscall.SIGHUP {
			echo("[syscall.SIGHUP].\r\n")
			// 割り込みを無視
			exit <- 1
		} else if s == syscall.SIGTERM {
			echo("[syscall.SIGTERM].\r\n")
			exit <- 2
		} else if s == os.Kill {
			echo("[os.Kill].\r\n")
			// 割り込みを無視
			exit <- 3
		} else if s == os.Interrupt {
			if runtime.GOOS != "darwin" {
				echo("[os.Interrupt].\r\n")
			}
			// 割り込みを無視
			exit <- 4
		} else if s == syscall.Signal(0x14) {
			if runtime.GOOS != "darwin" {
				echo("[syscall.SIGTSTP].\r\n")
			}
			// 割り込みを無視
			exit <- 5
		} else if s == syscall.SIGQUIT {
			echo("[syscall.SIGQUIT].\r\n")
			exit <- 6
		} else {
			// 未定義の割り込み処理
			echo("[Unknown syscall].\r\n")
			exit <- -1
		}
	}
}
