package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var (
	gopath = path.Join(os.Getenv("GOPATH"), "src") + "/"
)

func Err(err error) error {
	if err != nil {
		fmt.Printf("%s %s\n", location(2, true), err)
	}
	return err
}
func location(deep int, fullPath bool) string {
	_, file, line, ok := runtime.Caller(deep)
	if !ok {
		file = "???"
		line = 0
	}

	if fullPath {
		if strings.HasPrefix(file, gopath) {
			file = file[len(gopath):]
		}
	} else {
		file = filepath.Base(file)
	}
	return file + ":" + strconv.Itoa(line)
}
