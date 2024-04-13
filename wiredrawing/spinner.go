package wiredrawing

import (
	"fmt"
	"time"
)

func Spinner(c <-chan bool) {
	for {
		select {
		case isCompeled := <-c:
			if isCompeled == true {
				break
			}
		}
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(100 * time.Millisecond)
		}
	}
}
