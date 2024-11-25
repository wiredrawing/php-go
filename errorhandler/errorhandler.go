package errorhandler

import "log"

func Catch(args ...interface{}) interface{} {

	var err error = nil
	var ok bool
	err, ok = args[len(args)-1].(error)
	if ok == true {
		// ログを残して終了
		log.Fatalf("args: %v", args)
	}
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return err
}
