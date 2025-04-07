package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func spf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

func errTerm(err error) {
	// would be nice to identify where from error originated
	// some code inside err whould be huuuuge
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func allFiles(pref string, ext string) []string {
	entrySlice, err := os.ReadDir(pref)
	errTerm(err)
	var files []string
	suf := strings.HasSuffix
	for _, entry := range entrySlice {
		if !entry.IsDir() && suf(entry.Name(), ext) {
			files = append(files, entry.Name())
		}
	}
	return files
}

func sysCmd(argv []string) {

	run := exec.Command(argv[0], argv[1:]...)
	fmt.Println(run.Output())
}

func genGoFor(inProto string) {
	const gOut = "api/gen"
	sysCmd([]string{
		"protoc",
		spf("--go_out=%s ", gOut),
		spf("--go-grpc_out=%s ", gOut),
		inProto,
	})
}

func join(dir string, file string) string {
	return spf("%s/%s", dir, file)
}

func main() {
	// cmd := "protoc"
	protoPrefix := "api/proto"
	protos := allFiles(protoPrefix, ".proto")

	for _, fname := range protos {
		path := join(protoPrefix, fname)
		genGoFor(path)
	}

	// exec.Command()
}
