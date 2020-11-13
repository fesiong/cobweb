package cobweb

import (
	"os"
	"path/filepath"
	"unicode/utf8"
)

var ExecPath string

func initJSON() {
	sep := string(os.PathSeparator)
	root := filepath.Dir(os.Args[0])
	ExecPath, _ = filepath.Abs(root)
	length := utf8.RuneCountInString(ExecPath)
	lastChar := ExecPath[length-1:]
	if lastChar != sep {
		ExecPath = ExecPath + sep
	}
}

func init() {
	initJSON()
}
