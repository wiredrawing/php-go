package main

import (
	"fmt"
	"sync"

	"go-sample/wiredrawing"
	_ "go-sample/wiredrawing"

	"rsc.io/quote"
)

func main() {

	wiredrawing.Print()
	fmt.Println("test")
	var wg sync.WaitGroup
	fmt.Println(wg)
	fmt.Println(quote.Hello())

}
