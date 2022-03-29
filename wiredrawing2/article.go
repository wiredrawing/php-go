package wiredrawing2

// Article構造体
type Article struct {
	title       string
	description string
	author      string
}

// タイトルの設定
func (a *Article) SetTitle(title string) bool {

	// レシーバーにtitleを設定
	a.title = title
	return true
}

// 記事説明文
func (a *Article) SetDescription(description string) bool {

	a.description = description
	return true
}
