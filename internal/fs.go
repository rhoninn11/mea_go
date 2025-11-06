package internal

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const Fs_Path = "_fs"

func DirImage() string {
	return JoinPath(Fs_Path, "img")
}

func PathExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
func Filename(name string, ext string) string {
	return strings.Join([]string{name, ext}, ".")
}

func DirGuard(path string) {
	if !PathExist(path) {
		err := os.MkdirAll(path, 0775)
		if err != nil {
			log.Fatal(fmt.Errorf("for creating %s - %w", path, err))
		}
	}
}

func JoinPath(fragments ...string) string {
	return strings.Join(fragments, "/")
}
