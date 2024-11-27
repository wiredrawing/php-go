package config

func ColorWrapping(colorCode string, text string) string {
	return "\x1B[" + colorCode + "m" + text + "\x1B[0m"
}
