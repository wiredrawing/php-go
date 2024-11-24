package errorhandler

import "log"

func Catch(err error) {
	if err != nil {
		// ログを残して終了
		log.Fatal(err)
	}
}
