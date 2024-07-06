package errorhandler

import "log"

func ErrorHandler(err error) {
	if err != nil {
		// ログを残して終了
		log.Fatal(err)
	}
}
