# Go言語で初期化し開発準備をする



## 初期化コマンド

```
go mod init <something package name>


# go: creating new go.mod: <something package name>

```

上記のような内容が出力されて
go.mod というファイルが生成されていればOKです.

<something package name> => 作成するアプリケーションのトップレベルのパッケージ名を入力

## 初期化した状態で独自の任意のパッケージを作成する場合


somemodule (TOPディレクトリ)
  -> samplepackage
    -> index.go

上記のような階層にした場合

ソースコード上では

```

import (
  "somemodule/samplepackage"
)

// 上記のようなパス名にする

```
