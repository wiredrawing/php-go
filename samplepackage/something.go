package samplepackage

import "fmt"

// 外部から参照可能な関数
func CallableFunctionFromOtherPackage() {

	echo := fmt.Println
	fmt.Println("外部から参照可能なパッケージ関数")

	echo("fmt.Printlnメソッドをechoという変数に代入")
}
